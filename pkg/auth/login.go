package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LoginClient drives Zitadel's Session + OIDC auth-request API so the app can
// host its own sign-in UI instead of Zitadel's. It authenticates every call with
// an IAM_LOGIN_CLIENT service token, which MUST stay server-side — the SPA never
// sees it. Handlers depend on the LoginManager interface so they can be faked.
type LoginManager interface {
	// StartSession opens a session for loginName and reports whether the user has
	// a second factor configured (so the UI can prompt for it before finalizing).
	StartSession(ctx context.Context, loginName string) (Session, bool, error)
	// CheckPassword verifies a password against an existing session.
	CheckPassword(ctx context.Context, sessionID, sessionToken, password string) (Session, error)
	// CheckTOTP verifies a time-based one-time code against an existing session.
	CheckTOTP(ctx context.Context, sessionID, sessionToken, code string) (Session, error)
	// Finalize completes an OIDC auth request with a verified session and returns
	// the callback URL (carrying code + state) to redirect the browser to.
	Finalize(ctx context.Context, authRequestID, sessionID, sessionToken string) (callbackURL string, err error)
}

// Session is an in-flight Zitadel login session. SessionToken is bearer-grade
// for this session only and is returned to the SPA to carry between steps.
type Session struct {
	SessionID    string `json:"session_id"`
	SessionToken string `json:"session_token"`
}

// ZitadelLogin implements LoginManager against a Zitadel instance.
type ZitadelLogin struct {
	baseURL string
	token   TokenSource
	hc      *http.Client
}

// NewZitadelLogin builds a login client. baseURL is the issuer origin; token
// must resolve to an IAM_LOGIN_CLIENT-scoped service token.
func NewZitadelLogin(baseURL string, token TokenSource) *ZitadelLogin {
	return &ZitadelLogin{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		hc:      &http.Client{Timeout: 15 * time.Second},
	}
}

// call issues a JSON request with the login-client bearer token, decoding a
// non-nil out on success and surfacing the body on non-2xx.
func (l *ZitadelLogin) call(ctx context.Context, method, path string, body, out any) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("zitadel login: encode body: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, l.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("zitadel login: new request: %w", err)
	}
	tok, err := l.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")

	res, err := l.hc.Do(req)
	if err != nil {
		return fmt.Errorf("zitadel login: %s %s: %w", method, path, err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var b bytes.Buffer
		_, _ = b.ReadFrom(res.Body)
		return fmt.Errorf("zitadel login: %s %s: status %d: %s", method, path, res.StatusCode, b.String())
	}
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("zitadel login: decode response: %w", err)
		}
	}
	return nil
}

type sessionResp struct {
	SessionID    string `json:"sessionId"`
	SessionToken string `json:"sessionToken"`
}

// StartSession creates a session with a user (login-name) check and reports
// whether the resolved user has a TOTP second factor configured.
func (l *ZitadelLogin) StartSession(ctx context.Context, loginName string) (Session, bool, error) {
	body := map[string]any{
		"checks": map[string]any{
			"user": map[string]any{"loginName": loginName},
		},
	}
	var r sessionResp
	if err := l.call(ctx, http.MethodPost, "/v2/sessions", body, &r); err != nil {
		return Session{}, false, err
	}
	s := Session(r)
	mfa, err := l.userHasTOTP(ctx, r.SessionID)
	if err != nil {
		return Session{}, false, err
	}
	return s, mfa, nil
}

// userHasTOTP reports whether the session's user has a TOTP factor configured.
func (l *ZitadelLogin) userHasTOTP(ctx context.Context, sessionID string) (bool, error) {
	var sess struct {
		Session struct {
			Factors struct {
				User struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"factors"`
		} `json:"session"`
	}
	if err := l.call(ctx, http.MethodGet, "/v2/sessions/"+sessionID, nil, &sess); err != nil {
		return false, err
	}
	userID := sess.Session.Factors.User.ID
	if userID == "" {
		return false, nil
	}
	var methods struct {
		AuthMethodTypes []string `json:"authMethodTypes"`
	}
	if err := l.call(ctx, http.MethodGet,
		"/v2/users/"+userID+"/authentication_methods", nil, &methods); err != nil {
		return false, err
	}
	for _, m := range methods.AuthMethodTypes {
		if m == "AUTHENTICATION_METHOD_TYPE_TOTP" {
			return true, nil
		}
	}
	return false, nil
}

// CheckPassword adds a password check to a session, rotating its token.
func (l *ZitadelLogin) CheckPassword(ctx context.Context, sessionID, sessionToken, password string) (Session, error) {
	body := map[string]any{
		"sessionToken": sessionToken,
		"checks":       map[string]any{"password": map[string]any{"password": password}},
	}
	var r sessionResp
	if err := l.call(ctx, http.MethodPatch, "/v2/sessions/"+sessionID, body, &r); err != nil {
		return Session{}, err
	}
	tok := r.SessionToken
	if tok == "" {
		tok = sessionToken
	}
	return Session{SessionID: sessionID, SessionToken: tok}, nil
}

// CheckTOTP adds a TOTP (authenticator) check to a session, rotating its token.
func (l *ZitadelLogin) CheckTOTP(ctx context.Context, sessionID, sessionToken, code string) (Session, error) {
	body := map[string]any{
		"sessionToken": sessionToken,
		"checks":       map[string]any{"totp": map[string]any{"code": code}},
	}
	var r sessionResp
	if err := l.call(ctx, http.MethodPatch, "/v2/sessions/"+sessionID, body, &r); err != nil {
		return Session{}, err
	}
	tok := r.SessionToken
	if tok == "" {
		tok = sessionToken
	}
	return Session{SessionID: sessionID, SessionToken: tok}, nil
}

// Finalize completes the OIDC auth request and returns the browser callback URL.
func (l *ZitadelLogin) Finalize(ctx context.Context, authRequestID, sessionID, sessionToken string) (string, error) {
	body := map[string]any{
		"session": map[string]any{"sessionId": sessionID, "sessionToken": sessionToken},
	}
	var r struct {
		CallbackURL string `json:"callbackUrl"`
	}
	if err := l.call(ctx, http.MethodPost,
		"/v2/oidc/auth_requests/"+authRequestID, body, &r); err != nil {
		return "", err
	}
	if r.CallbackURL == "" {
		return "", fmt.Errorf("zitadel login: finalize returned empty callback url")
	}
	return r.CallbackURL, nil
}
