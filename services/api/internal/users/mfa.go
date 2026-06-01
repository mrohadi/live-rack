package users

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
)

// Enroller is the Zitadel surface authenticator enrollment needs.
// *auth.ZitadelManagement satisfies it.
type Enroller interface {
	RegisterTOTP(ctx context.Context, userID string) (uri, secret string, err error)
	VerifyTOTP(ctx context.Context, userID, code string) error
}

// MFAStore records a user's verified second factor. *store.Queries satisfies it.
type MFAStore interface {
	SetUserMFA(ctx context.Context, userID, orgID uuid.UUID, enabled bool) error
}

// MFAHandler serves self-service authenticator (TOTP) enrollment for the caller.
type MFAHandler struct {
	zit   Enroller
	store MFAStore
	audit Auditor
}

// NewMFA builds an MFAHandler.
func NewMFA(zit Enroller, s MFAStore, a Auditor) *MFAHandler {
	return &MFAHandler{zit: zit, store: s, audit: a}
}

// Register mounts enrollment routes on the authenticated API group.
func (h *MFAHandler) Register(g *echo.Group) {
	g.POST("/me/2fa/totp", h.Start)
	g.POST("/me/2fa/totp/verify", h.Verify)
}

// StartTOTPResponse carries the provisioning data for an authenticator app.
type StartTOTPResponse struct {
	URI    string `json:"uri"`
	Secret string `json:"secret"`
}

// Start godoc
//
//	@Summary	Begin authenticator (TOTP) enrollment for the caller
//	@Tags		users
//	@Produce	json
//	@Success	200	{object}	StartTOTPResponse
//	@Router		/me/2fa/totp [post]
func (h *MFAHandler) Start(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if p.IDPUserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "no idp user")
	}
	uri, secret, err := h.zit.RegisterTOTP(c.Request().Context(), p.IDPUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "register totp")
	}
	return c.JSON(http.StatusOK, StartTOTPResponse{URI: uri, Secret: secret})
}

// VerifyTOTPRequest carries the first code from the authenticator app.
type VerifyTOTPRequest struct {
	Code string `json:"code"`
}

// Verify godoc
//
//	@Summary	Confirm authenticator enrollment with the first code
//	@Tags		users
//	@Accept		json
//	@Param		body	body	VerifyTOTPRequest	true	"TOTP code"
//	@Success	204
//	@Router		/me/2fa/totp/verify [post]
func (h *MFAHandler) Verify(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if p.IDPUserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "no idp user")
	}
	var req VerifyTOTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing code")
	}
	ctx := c.Request().Context()
	if err := h.zit.VerifyTOTP(ctx, p.IDPUserID, code); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid code")
	}
	// Record coverage so the roster + stats reflect the new second factor.
	if err := h.store.SetUserMFA(ctx, p.UserID, p.OrgID, true); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "record mfa")
	}
	_ = h.audit.Write(ctx, audit.Entry{
		OrgID:        p.OrgID,
		ActorUserID:  p.UserID,
		Action:       "user.2fa_enrolled",
		ResourceType: "user",
		ResourceID:   p.IDPUserID,
	})
	return c.NoContent(http.StatusNoContent)
}
