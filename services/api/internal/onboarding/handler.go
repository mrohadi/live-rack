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
)

// Completer is the Zitadel surface invite acceptance needs.
// *auth.ZitadelManagement satisfies it.
type Completer interface {
	VerifyEmail(ctx context.Context, userID, code string) error
	SetPassword(ctx context.Context, orgID, userID, password string) error
}

// Handler serves the public invite-acceptance endpoint.
type Handler struct {
	zit Completer
}

// New builds an onboarding Handler.
func New(z Completer) *Handler {
	return &Handler{zit: z}
}

// Register mounts the public onboarding routes on the root router (no auth).
func (h *Handler) Register(e *echo.Echo) {
	e.POST("/api/v1/onboard/complete", h.Complete)
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
