// Package onboarding completes invite acceptance from the app's own screen:
// verify the email with the code from the invite link, then set the user's
// password. Routes are public (pre-authentication); the management service
// token stays server-side. Zitadel enforces the password policy.
package onboarding

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
)

// Completer is the Zitadel surface invite acceptance needs.
// *auth.ZitadelManagement satisfies it.
type Completer interface {
	VerifyEmail(ctx context.Context, userID, code string) error
	SetPassword(ctx context.Context, orgID, userID, password string) error
	GetLoginName(ctx context.Context, userID string) (string, error)
	RegisterTOTP(ctx context.Context, userID string) (uri, secret string, err error)
	VerifyTOTP(ctx context.Context, userID, code string) error
}

// PasswordChecker validates a password by opening a Zitadel session — used to
// gate TOTP enrollment so only the new user (who just set the password) can
// enroll. *auth.ZitadelLogin satisfies it.
type PasswordChecker interface {
	StartSession(ctx context.Context, loginName string) (pkgauth.Session, bool, error)
	CheckPassword(ctx context.Context, sessionID, sessionToken, password string) (pkgauth.Session, error)
}

// Handler serves the public invite-acceptance endpoints.
type Handler struct {
	zit   Completer
	login PasswordChecker
}

// New builds an onboarding Handler.
func New(z Completer, login PasswordChecker) *Handler {
	return &Handler{zit: z, login: login}
}

// Register mounts the public onboarding routes on the root router (no auth).
func (h *Handler) Register(e *echo.Echo) {
	e.POST("/api/v1/onboard/complete", h.Complete)
	e.POST("/api/v1/onboard/totp/start", h.TOTPStart)
	e.POST("/api/v1/onboard/totp/verify", h.TOTPVerify)
}

// gate re-validates the just-set password before any enrollment action, so a
// stranger cannot enroll a factor on someone else's freshly-invited account.
func (h *Handler) gate(ctx context.Context, userID, password string) error {
	if userID == "" || password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing user or password")
	}
	loginName, err := h.zit.GetLoginName(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "unknown user")
	}
	sess, _, err := h.login.StartSession(ctx, loginName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "session")
	}
	if _, err := h.login.CheckPassword(ctx, sess.SessionID, sess.SessionToken, password); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid password")
	}
	return nil
}

// CompleteRequest carries the invite-link params plus the chosen password.
type CompleteRequest struct {
	UserID   string `json:"user_id"`
	OrgID    string `json:"org_id"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

// Complete godoc
//
//	@Summary	Accept an invite: verify email + set password
//	@Tags		onboarding
//	@Accept		json
//	@Param		body	body	CompleteRequest	true	"Invite acceptance"
//	@Success	204
//	@Router		/onboard/complete [post]
func (h *Handler) Complete(c echo.Context) error {
	var req CompleteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	userID := strings.TrimSpace(req.UserID)
	orgID := strings.TrimSpace(req.OrgID)
	code := strings.TrimSpace(req.Code)
	if userID == "" || orgID == "" || code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing user, org, or code")
	}
	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "password too short")
	}
	ctx := c.Request().Context()
	if err := h.zit.VerifyEmail(ctx, userID, code); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid or expired code")
	}
	if err := h.zit.SetPassword(ctx, orgID, userID, req.Password); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "password rejected")
	}
	return c.NoContent(http.StatusNoContent)
}

// TOTPStartRequest gates enrollment with the just-set password.
type TOTPStartRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

// TOTPStartResponse carries the authenticator provisioning data.
type TOTPStartResponse struct {
	URI    string `json:"uri"`
	Secret string `json:"secret"`
}

// TOTPStart godoc
//
//	@Summary	Begin authenticator enrollment during invite onboarding
//	@Tags		onboarding
//	@Accept		json
//	@Produce	json
//	@Param		body	body		TOTPStartRequest	true	"User + password"
//	@Success	200		{object}	TOTPStartResponse
//	@Router		/onboard/totp/start [post]
func (h *Handler) TOTPStart(c echo.Context) error {
	var req TOTPStartRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	userID := strings.TrimSpace(req.UserID)
	ctx := c.Request().Context()
	if err := h.gate(ctx, userID, req.Password); err != nil {
		return err
	}
	uri, secret, err := h.zit.RegisterTOTP(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "register totp")
	}
	return c.JSON(http.StatusOK, TOTPStartResponse{URI: uri, Secret: secret})
}

// TOTPVerifyRequest confirms enrollment with the first code (password-gated).
type TOTPVerifyRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// TOTPVerify godoc
//
//	@Summary	Confirm authenticator enrollment during invite onboarding
//	@Tags		onboarding
//	@Accept		json
//	@Param		body	body	TOTPVerifyRequest	true	"User + password + code"
//	@Success	204
//	@Router		/onboard/totp/verify [post]
func (h *Handler) TOTPVerify(c echo.Context) error {
	var req TOTPVerifyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	userID := strings.TrimSpace(req.UserID)
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing code")
	}
	ctx := c.Request().Context()
	if err := h.gate(ctx, userID, req.Password); err != nil {
		return err
	}
	if err := h.zit.VerifyTOTP(ctx, userID, code); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid code")
	}
	return c.NoContent(http.StatusNoContent)
}
