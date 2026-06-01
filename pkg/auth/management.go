package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
}

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

// GrantProjectRole grants the configured project's role to a user.
func (m *ZitadelManagement) GrantProjectRole(ctx context.Context, orgID, userID, role string) error {
	reqBody := map[string]any{
		"projectId": m.projectID,
		"roleKeys":  []string{role},
	}
	return m.post(ctx,
		fmt.Sprintf("/management/v1/users/%s/grants", userID), orgID, reqBody, nil)
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
	return resp.URI, resp.Secret, nil
}

// VerifyTOTP confirms authenticator enrollment with the user's first code.
func (m *ZitadelManagement) VerifyTOTP(ctx context.Context, userID, code string) error {
	return m.post(ctx, fmt.Sprintf("/v2/users/%s/totp/verify", userID), "",
		map[string]any{"code": code}, nil)
}
