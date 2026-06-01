// Package passwordreset serves the public forgot/reset-password flow backed by
// Zitadel's v2 user API. Routes are pre-authentication; the management service
// token stays server-side. The forgot endpoint never reveals whether an email
// is registered (no user enumeration).
package passwordreset

import (
	"context"
	"net/http"
	"net/mail"
	"strings"

	"github.com/labstack/echo/v4"
)

// Resetter is the Zitadel surface the flow needs.
// *auth.ZitadelManagement satisfies it.
type Resetter interface {
	FindUserByEmail(ctx context.Context, email string) (string, error)
	SendPasswordResetCode(ctx context.Context, userID string) error
	ResetPassword(ctx context.Context, userID, code, password string) error
}

// Handler serves the public forgot/reset-password endpoints.
type Handler struct {
	zit Resetter
}

// New builds a Handler.
func New(z Resetter) *Handler {
	return &Handler{zit: z}
}

// Register mounts the public routes on the root router (no auth middleware).
func (h *Handler) Register(e *echo.Echo) {
	e.POST("/api/v1/password/forgot", h.Forgot)
	e.POST("/api/v1/password/reset", h.Reset)
}

// ForgotRequest names the account to email a reset link to.
type ForgotRequest struct {
	Email string `json:"email"`
}

// Forgot godoc
//
//	@Summary	Email a password-reset link (no user enumeration)
//	@Tags		password
//	@Accept		json
//	@Param		body	body	ForgotRequest	true	"Email"
//	@Success	204
//	@Router		/password/forgot [post]
func (h *Handler) Forgot(c echo.Context) error {
	var req ForgotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email")
	}
	ctx := c.Request().Context()
	// Best-effort: resolve + send, but always return 204 so the response does
	// not reveal whether the address is registered.
	if userID, err := h.zit.FindUserByEmail(ctx, email); err == nil && userID != "" {
		_ = h.zit.SendPasswordResetCode(ctx, userID)
	}
	return c.NoContent(http.StatusNoContent)
}

// ResetRequest carries the reset-link params plus the chosen password.
type ResetRequest struct {
	UserID   string `json:"user_id"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

// Reset godoc
//
//	@Summary	Set a new password with a reset code
//	@Tags		password
//	@Accept		json
//	@Param		body	body	ResetRequest	true	"Reset"
//	@Success	204
//	@Router		/password/reset [post]
func (h *Handler) Reset(c echo.Context) error {
	var req ResetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	userID := strings.TrimSpace(req.UserID)
	code := strings.TrimSpace(req.Code)
	if userID == "" || code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing user or code")
	}
	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "password too short")
	}
	if err := h.zit.ResetPassword(c.Request().Context(), userID, code, req.Password); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid or expired code")
	}
	return c.NoContent(http.StatusNoContent)
}
