package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Management is the slice of Zitadel's admin/management API the app drives for
// onboarding: creating tenant orgs, inviting users, and granting project roles.
// Handlers depend on this interface so they can be faked in tests.
type Management interface {
	// CreateOrg provisions a new tenant organization and returns its Zitadel id.
	CreateOrg(ctx context.Context, name string) (orgID string, err error)
	// CreateHumanUser creates a human user inside orgID and triggers an invite /
	// initialization email so they can set their own password + second factor.
	CreateHumanUser(ctx context.Context, orgID, email, displayName string) (userID string, err error)
	// CreateHumanUserReturnCode creates a human user and returns the email
	// verification code directly instead of sending it via email. Use in
	// development so the signup flow works without SMTP delivery.
	CreateHumanUserReturnCode(ctx context.Context, orgID, email, displayName string) (userID, code string, err error)
	// GrantProjectRole grants the configured project's role to a user in orgID.
	GrantProjectRole(ctx context.Context, orgID, userID, role string) error
	// ResendInvite re-sends the initialization email for a not-yet-active user.
	ResendInvite(ctx context.Context, orgID, userID string) error
	// PendingInvites counts users still in the initial (not-yet-verified) state.
	PendingInvites(ctx context.Context, orgID string) (int, error)
	// SendPasswordReset emails a user a password-reset link.
	SendPasswordReset(ctx context.Context, orgID, userID string) error
	// RegisterTOTP starts authenticator enrollment for a user, returning the
	// otpauth:// provisioning URI (for a QR code) and the shared secret.
	RegisterTOTP(ctx context.Context, userID string) (uri, secret string, err error)
	// VerifyTOTP confirms enrollment by validating the user's first code.
	VerifyTOTP(ctx context.Context, userID, code string) error
	// VerifyEmail confirms a user's email with the code from their invite email.
	VerifyEmail(ctx context.Context, userID, code string) error
	// SetPassword sets a user's password (admin authority, no current password).
	SetPassword(ctx context.Context, orgID, userID, password string) error
	// GetLoginName returns a user's preferred login name.
	GetLoginName(ctx context.Context, userID string) (string, error)
	// FindUserByEmail resolves a user id from an email, or "" if none exists.
	FindUserByEmail(ctx context.Context, email string) (string, error)
	// SendPasswordResetCode emails a user a reset code linking to our own
	// reset-password screen.
	SendPasswordResetCode(ctx context.Context, userID string) error
	// ResetPassword sets a new password using a reset verification code.
	ResetPassword(ctx context.Context, userID, code, password string) error
}

// TokenSource yields a bearer token authorized for Zitadel management calls.
// Concrete deployments back this with a service-account (JWT profile / PAT).
type TokenSource func(ctx context.Context) (string, error)

// StaticToken adapts a fixed token (e.g. a service-account PAT) to a TokenSource.
func StaticToken(tok string) TokenSource {
	return func(context.Context) (string, error) {
		if tok == "" {
			return "", fmt.Errorf("zitadel: empty management token")
		}
		return tok, nil
	}
}

// ZitadelManagement talks to a Zitadel instance's REST API. Endpoints pinned to
// the v2.71 surface; org context for management v1 calls rides the
// x-zitadel-orgid header.
type ZitadelManagement struct {
	baseURL    string
	projectID  string
	appBaseURL string
	token      TokenSource
	hc         *http.Client

	ownerMu  sync.Mutex
	ownerOrg string // resolved org that owns projectID (cached)
}

// projectRoleKeys are every role the live-rack project defines. A tenant org
// receives all of them via its project grant; users are then granted a subset.
var projectRoleKeys = []string{"admin", "manager", "staff", "readonly", "service"}

// NewZitadelManagement builds a management client. baseURL is the issuer origin
// (e.g. http://localhost:8081); projectID scopes role grants; appBaseURL is the
// SPA origin that invite emails link back to (e.g. http://localhost:5173).
func NewZitadelManagement(baseURL, projectID, appBaseURL string, token TokenSource) *ZitadelManagement {
	return &ZitadelManagement{
		baseURL:    strings.TrimRight(baseURL, "/"),
		projectID:  projectID,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
		token:      token,
		hc:         &http.Client{Timeout: 15 * time.Second},
	}
}

// post issues a JSON POST, optionally scoped to orgID, decoding the response
// into out when non-nil. Non-2xx responses become errors carrying the body.
func (m *ZitadelManagement) post(ctx context.Context, path, orgID string, body, out any) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("zitadel: encode body: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("zitadel: new request: %w", err)
	}
	tok, err := m.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	if orgID != "" {
		req.Header.Set("x-zitadel-orgid", orgID)
	}

	res, err := m.hc.Do(req)
	if err != nil {
		return fmt.Errorf("zitadel: POST %s: %w", path, err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var b bytes.Buffer
		_, _ = b.ReadFrom(res.Body)
		return fmt.Errorf("zitadel: POST %s: status %d: %s", path, res.StatusCode, b.String())
	}
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("zitadel: decode response: %w", err)
		}
	}
	return nil
}

// get issues a JSON GET with the management bearer token, decoding into out.
func (m *ZitadelManagement) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("zitadel: new request: %w", err)
	}
	tok, err := m.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)

	res, err := m.hc.Do(req)
	if err != nil {
		return fmt.Errorf("zitadel: GET %s: %w", path, err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var b bytes.Buffer
		_, _ = b.ReadFrom(res.Body)
		return fmt.Errorf("zitadel: GET %s: status %d: %s", path, res.StatusCode, b.String())
	}
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("zitadel: decode response: %w", err)
		}
	}
	return nil
}

// GetLoginName returns a user's preferred login name (used to gate onboarding
// actions by re-validating the just-set password).
func (m *ZitadelManagement) GetLoginName(ctx context.Context, userID string) (string, error) {
	var resp struct {
		User struct {
			PreferredLoginName string `json:"preferredLoginName"`
		} `json:"user"`
	}
	if err := m.get(ctx, fmt.Sprintf("/v2/users/%s", userID), &resp); err != nil {
		return "", err
	}
	if resp.User.PreferredLoginName == "" {
		return "", fmt.Errorf("zitadel: user %s has no login name", userID)
	}
	return resp.User.PreferredLoginName, nil
}

// CreateOrg provisions a tenant org via the management API.
func (m *ZitadelManagement) CreateOrg(ctx context.Context, name string) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := m.post(ctx, "/management/v1/orgs", "",
		map[string]string{"name": name}, &resp); err != nil {
		return "", err
	}
	if resp.ID == "" {
		return "", fmt.Errorf("zitadel: create org returned empty id")
	}
	return resp.ID, nil
}

// splitName splits a display name into given/family parts for Zitadel's profile.
// Pure: single-token names reuse the token as the family name.
func splitName(display string) (given, family string) {
	parts := strings.Fields(strings.TrimSpace(display))
	switch len(parts) {
	case 0:
		return "Member", "Member"
	case 1:
		return parts[0], parts[0]
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}

// CreateHumanUser creates a human user in orgID and requests an email-verified
// invite flow so the invitee sets their own credentials.
func (m *ZitadelManagement) CreateHumanUser(ctx context.Context, orgID, email, displayName string) (string, error) {
	given, family := splitName(displayName)
	// Link the verification email to our own onboarding screen instead of the
	// Zitadel hosted page. Zitadel fills {{.Code}}/{{.UserID}}/{{.OrgID}}.
	urlTemplate := m.appBaseURL + "/verify-email?code={{.Code}}&userID={{.UserID}}&orgID={{.OrgID}}"
	reqBody := map[string]any{
		"organization": map[string]string{"orgId": orgID},
		"profile":      map[string]string{"givenName": given, "familyName": family},
		"email": map[string]any{
			"email":    email,
			"sendCode": map[string]any{"urlTemplate": urlTemplate},
		},
	}
	var resp struct {
		UserID string `json:"userId"`
	}
	if err := m.post(ctx, "/v2/users/human", orgID, reqBody, &resp); err != nil {
		return "", err
	}
	if resp.UserID == "" {
		return "", fmt.Errorf("zitadel: create user returned empty id")
	}
	return resp.UserID, nil
}

// CreateHumanUserReturnCode creates a user and returns the email verification
// code directly (Zitadel returnCode mode). Use only in development — no email
// is sent; the caller is responsible for presenting the code to the user.
func (m *ZitadelManagement) CreateHumanUserReturnCode(ctx context.Context, orgID, email, displayName string) (string, string, error) {
	given, family := splitName(displayName)
	reqBody := map[string]any{
		"organization": map[string]string{"orgId": orgID},
		"profile":      map[string]string{"givenName": given, "familyName": family},
		"email": map[string]any{
			"email":      email,
			"returnCode": map[string]any{},
		},
	}
	var resp struct {
		UserID    string `json:"userId"`
		EmailCode string `json:"emailCode"`
	}
	if err := m.post(ctx, "/v2/users/human", orgID, reqBody, &resp); err != nil {
		return "", "", err
	}
	if resp.UserID == "" {
		return "", "", fmt.Errorf("zitadel: create user returned empty id")
	}
	return resp.UserID, resp.EmailCode, nil
}

// GrantProjectRole grants the configured project's role to a user.
//
// The live-rack project is owned by one org. Users in that owner org get a
// direct user grant. Users in a *tenant* org (created by self-service signup)
// cannot be granted a project they do not own — Zitadel returns "Project not
// found" — so we first ensure the project is granted to that tenant org, then
// reference the resulting project grant on the user grant.
func (m *ZitadelManagement) GrantProjectRole(ctx context.Context, orgID, userID, role string) error {
	owner, err := m.projectOwnerOrg(ctx)
	if err != nil {
		return err
	}

	reqBody := map[string]any{
		"projectId": m.projectID,
		"roleKeys":  []string{role},
	}
	if orgID != owner {
		grantID, gerr := m.ensureProjectGrant(ctx, owner, orgID)
		if gerr != nil {
			return gerr
		}
		reqBody["projectGrantId"] = grantID
	}

	return m.post(ctx,
		fmt.Sprintf("/management/v1/users/%s/grants", userID), orgID, reqBody, nil)
}

// projectOwnerOrg returns the org that owns the configured project, caching it
// after the first lookup.
func (m *ZitadelManagement) projectOwnerOrg(ctx context.Context) (string, error) {
	m.ownerMu.Lock()
	defer m.ownerMu.Unlock()
	if m.ownerOrg != "" {
		return m.ownerOrg, nil
	}
	var resp struct {
		Project struct {
			Details struct {
				ResourceOwner string `json:"resourceOwner"`
			} `json:"details"`
		} `json:"project"`
	}
	if err := m.get(ctx, "/management/v1/projects/"+m.projectID, &resp); err != nil {
		return "", err
	}
	owner := resp.Project.Details.ResourceOwner
	if owner == "" {
		return "", fmt.Errorf("zitadel: project %s has no resource owner", m.projectID)
	}
	m.ownerOrg = owner
	return owner, nil
}

// ensureProjectGrant returns the id of the project grant that shares the project
// with grantedOrg, creating it (with every project role) if absent. Idempotent.
func (m *ZitadelManagement) ensureProjectGrant(ctx context.Context, ownerOrg, grantedOrg string) (string, error) {
	var search struct {
		Result []struct {
			GrantID      string `json:"grantId"`
			GrantedOrgID string `json:"grantedOrgId"`
		} `json:"result"`
	}
	if err := m.post(ctx,
		fmt.Sprintf("/management/v1/projects/%s/grants/_search", m.projectID),
		ownerOrg, map[string]any{}, &search); err != nil {
		return "", err
	}
	for _, g := range search.Result {
		if g.GrantedOrgID == grantedOrg {
			return g.GrantID, nil
		}
	}

	var created struct {
		GrantID string `json:"grantId"`
	}
	if err := m.post(ctx,
		fmt.Sprintf("/management/v1/projects/%s/grants", m.projectID),
		ownerOrg, map[string]any{
			"grantedOrgId": grantedOrg,
			"roleKeys":     projectRoleKeys,
		}, &created); err != nil {
		return "", err
	}
	if created.GrantID == "" {
		return "", fmt.Errorf("zitadel: project grant to org %s returned empty id", grantedOrg)
	}
	return created.GrantID, nil
}

// ResendInvite re-sends a user's initialization email.
func (m *ZitadelManagement) ResendInvite(ctx context.Context, orgID, userID string) error {
	return m.post(ctx,
		fmt.Sprintf("/management/v1/users/%s/_resend_initialization", userID), orgID,
		map[string]any{}, nil)
}

// PendingInvites counts org users still in USER_STATE_INITIAL (invited, not yet
// verified). Returns 0 on a missing/!2xx response so the UI degrades gracefully.
func (m *ZitadelManagement) PendingInvites(ctx context.Context, orgID string) (int, error) {
	body := map[string]any{
		"queries": []any{
			map[string]any{"stateQuery": map[string]any{"state": "USER_STATE_INITIAL"}},
		},
	}
	var resp struct {
		Details struct {
			TotalResult string `json:"totalResult"`
		} `json:"details"`
	}
	if err := m.post(ctx, "/management/v1/users/_search", orgID, body, &resp); err != nil {
		return 0, err
	}
	n, _ := strconv.Atoi(resp.Details.TotalResult)
	return n, nil
}

// SendPasswordReset asks Zitadel to email the user a password-reset link.
func (m *ZitadelManagement) SendPasswordReset(ctx context.Context, orgID, userID string) error {
	return m.post(ctx,
		fmt.Sprintf("/management/v1/users/%s/_reset_password", userID), orgID,
		map[string]any{}, nil)
}

// FindUserByEmail resolves a user id from an email via the v2 search endpoint.
// Returns "" (no error) when no user matches, so callers can avoid leaking
// whether an address is registered.
func (m *ZitadelManagement) FindUserByEmail(ctx context.Context, email string) (string, error) {
	body := map[string]any{
		"queries": []any{
			map[string]any{"emailQuery": map[string]any{"emailAddress": email}},
		},
	}
	var resp struct {
		Result []struct {
			UserID string `json:"userId"`
		} `json:"result"`
	}
	if err := m.post(ctx, "/v2/users", "", body, &resp); err != nil {
		return "", err
	}
	if len(resp.Result) == 0 {
		return "", nil
	}
	return resp.Result[0].UserID, nil
}

// SendPasswordResetCode emails a reset code linking to our reset-password screen.
func (m *ZitadelManagement) SendPasswordResetCode(ctx context.Context, userID string) error {
	urlTemplate := m.appBaseURL + "/reset-password?code={{.Code}}&userID={{.UserID}}"
	return m.post(ctx, fmt.Sprintf("/v2/users/%s/password_reset", userID), "",
		map[string]any{"sendLink": map[string]any{"urlTemplate": urlTemplate}}, nil)
}

// ResetPassword sets a new password using the reset verification code.
func (m *ZitadelManagement) ResetPassword(ctx context.Context, userID, code, password string) error {
	return m.post(ctx, fmt.Sprintf("/v2/users/%s/password", userID), "",
		map[string]any{
			"newPassword":      map[string]any{"password": password},
			"verificationCode": code,
		}, nil)
}

// VerifyEmail confirms a user's email address with the code from their invite
// email (v2 user service).
func (m *ZitadelManagement) VerifyEmail(ctx context.Context, userID, code string) error {
	return m.post(ctx, fmt.Sprintf("/v2/users/%s/email/verify", userID), "",
		map[string]any{"verificationCode": code}, nil)
}

// SetPassword sets a user's password with admin authority (no current password
// required) via the management API; used to complete invite onboarding.
func (m *ZitadelManagement) SetPassword(ctx context.Context, orgID, userID, password string) error {
	return m.post(ctx, fmt.Sprintf("/management/v1/users/%s/password", userID), orgID,
		map[string]any{"password": password, "noChangeRequired": true}, nil)
}

// RegisterTOTP starts authenticator enrollment via the v2 user service. The
// returned uri encodes the otpauth:// provisioning string for a QR code; secret
// is the manual-entry fallback. Enrollment is not active until VerifyTOTP.
// The otpauth issuer is rewritten from the Zitadel instance name to "live-rack"
// so authenticator apps show the correct label.
func (m *ZitadelManagement) RegisterTOTP(ctx context.Context, userID string) (string, string, error) {
	var resp struct {
		URI    string `json:"uri"`
		Secret string `json:"secret"`
	}
	if err := m.post(ctx, fmt.Sprintf("/v2/users/%s/totp", userID), "",
		map[string]any{}, &resp); err != nil {
		return "", "", err
	}
	if resp.URI == "" || resp.Secret == "" {
		return "", "", fmt.Errorf("zitadel: register totp returned empty uri/secret")
	}
	return rewriteTOTPIssuer(resp.URI, "live-rack"), resp.Secret, nil
}

// rewriteTOTPIssuer replaces the issuer segment in an otpauth:// URI.
// Format: otpauth://totp/<issuer>:<account>?issuer=<issuer>&...
func rewriteTOTPIssuer(rawURI, newIssuer string) string {
	u, err := url.Parse(rawURI)
	if err != nil {
		return rawURI
	}
	// The path is "/<issuer>:<account>" — replace the issuer prefix.
	p := strings.TrimPrefix(u.Path, "/")
	if idx := strings.Index(p, ":"); idx >= 0 {
		account := p[idx+1:]
		u.Path = "/" + url.PathEscape(newIssuer) + ":" + account
	}
	// Also fix the issuer query param and remove any encoded slash artefacts.
	q := u.Query()
	if q.Get("issuer") != "" {
		q.Set("issuer", newIssuer)
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// VerifyTOTP confirms authenticator enrollment with the user's first code.
func (m *ZitadelManagement) VerifyTOTP(ctx context.Context, userID, code string) error {
	return m.post(ctx, fmt.Sprintf("/v2/users/%s/totp/verify", userID), "",
		map[string]any{"code": code}, nil)
}
