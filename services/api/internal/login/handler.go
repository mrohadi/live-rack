// Package login proxies Zitadel's Session + OIDC auth-request API so the app can
// host its own sign-in UI. Routes are public (pre-authentication); the
// IAM_LOGIN_CLIENT service token stays server-side and is never exposed to the
// SPA. The SPA carries the opaque session id + token between steps.
package login

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
)

// Handler serves the public custom-login endpoints.
type Handler struct {
	login pkgauth.LoginManager
}

// New builds a login Handler.
func New(l pkgauth.LoginManager) *Handler {
	return &Handler{login: l}
}

// Register mounts the public login routes on the root router (no auth middleware).
func (h *Handler) Register(e *echo.Echo) {
	g := e.Group("/api/v1/login")
	g.POST("/start", h.Start)
	g.POST("/password", h.Password)
	g.POST("/totp", h.TOTP)
	g.POST("/finalize", h.Finalize)
}

// sessionResponse echoes the opaque session handles the SPA carries forward.
type sessionResponse struct {
	SessionID    string `json:"session_id"`
	SessionToken string `json:"session_token"`
}

// startResponse adds the MFA hint so the UI can prompt for a code before finalizing.
type startResponse struct {
	SessionID    string `json:"session_id"`
	SessionToken string `json:"session_token"`
	MFARequired  bool   `json:"mfa_required"`
}

// StartRequest names the user to begin a login session for.
type StartRequest struct {
	LoginName string `json:"login_name"`
}

// Start opens a session for a login name.
func (h *Handler) Start(c echo.Context) error {
	var req StartRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	name := strings.TrimSpace(strings.ToLower(req.LoginName))
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "login_name required")
	}
	s, mfa, err := h.login.StartSession(c.Request().Context(), name)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unknown user")
	}
	return c.JSON(http.StatusOK, startResponse{
		SessionID: s.SessionID, SessionToken: s.SessionToken, MFARequired: mfa,
	})
}

// PasswordRequest carries a session and the password to check against it.
type PasswordRequest struct {
	SessionID    string `json:"session_id"`
	SessionToken string `json:"session_token"`
	Password     string `json:"password"`
}

// Password verifies a password for an open session.
func (h *Handler) Password(c echo.Context) error {
	var req PasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.SessionID == "" || req.SessionToken == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "session and password required")
	}
	s, err := h.login.CheckPassword(c.Request().Context(), req.SessionID, req.SessionToken, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	return c.JSON(http.StatusOK, sessionResponse{SessionID: s.SessionID, SessionToken: s.SessionToken})
}

// TOTPRequest carries a session and a one-time authenticator code.
type TOTPRequest struct {
	SessionID    string `json:"session_id"`
	SessionToken string `json:"session_token"`
	Code         string `json:"code"`
}

// TOTP verifies a second-factor authenticator code for an open session.
func (h *Handler) TOTP(c echo.Context) error {
	var req TOTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.SessionID == "" || req.SessionToken == "" || req.Code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "session and code required")
	}
	s, err := h.login.CheckTOTP(c.Request().Context(), req.SessionID, req.SessionToken, req.Code)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid code")
	}
	return c.JSON(http.StatusOK, sessionResponse{SessionID: s.SessionID, SessionToken: s.SessionToken})
}

// FinalizeRequest binds a verified session to an OIDC auth request.
type FinalizeRequest struct {
	AuthRequestID string `json:"auth_request_id"`
	SessionID     string `json:"session_id"`
	SessionToken  string `json:"session_token"`
}

// FinalizeResponse returns the browser redirect target carrying code + state.
type FinalizeResponse struct {
	CallbackURL string `json:"callback_url"`
}

// Finalize completes the OIDC auth request and returns the callback URL.
func (h *Handler) Finalize(c echo.Context) error {
	var req FinalizeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.AuthRequestID == "" || req.SessionID == "" || req.SessionToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "auth request and session required")
	}
	url, err := finalize(c.Request().Context(), h.login, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "finalize auth request")
	}
	return c.JSON(http.StatusOK, FinalizeResponse{CallbackURL: url})
}

func finalize(ctx context.Context, l pkgauth.LoginManager, req FinalizeRequest) (string, error) {
	return l.Finalize(ctx, req.AuthRequestID, req.SessionID, req.SessionToken)
}
