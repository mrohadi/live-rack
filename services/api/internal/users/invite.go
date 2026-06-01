package users

import (
	"context"
	"net/http"
	"net/mail"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
)

// Inviter is the Zitadel management surface invites need. *auth.ZitadelManagement
// satisfies it.
type Inviter interface {
	CreateHumanUser(ctx context.Context, orgID, email, displayName string) (string, error)
	GrantProjectRole(ctx context.Context, orgID, userID, role string) error
	ResendInvite(ctx context.Context, orgID, userID string) error
}

// Auditor records an append-only audit entry. *audit.Writer satisfies it.
type Auditor interface {
	Write(ctx context.Context, e audit.Entry) error
}

// assignableRoles are the roles an admin may grant via invite. "service" is
// reserved for service tokens and never invited.
var assignableRoles = map[string]bool{
	string(domain.RoleAdmin):    true,
	string(domain.RoleManager):  true,
	string(domain.RoleStaff):    true,
	string(domain.RoleReadonly): true,
}

// InviteHandler serves admin-only user onboarding endpoints.
type InviteHandler struct {
	zit   Inviter
	audit Auditor
}

// NewInvite builds an InviteHandler.
func NewInvite(zit Inviter, a Auditor) *InviteHandler {
	return &InviteHandler{zit: zit, audit: a}
}

// Register mounts invite routes on the authenticated API group.
func (h *InviteHandler) Register(g *echo.Group) {
	g.POST("/users/invite", h.Invite)
	g.POST("/users/:id/resend", h.Resend)
}

// InviteRequest names a teammate to invite and the role to grant.
type InviteRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

// InviteResponse echoes the pending invitee.
type InviteResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// Invite godoc
//
//	@Summary	Invite a user to the org (admin + 2FA required)
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		body	body		InviteRequest	true	"Invitee"
//	@Success	201		{object}	InviteResponse
//	@Router		/users/invite [post]
func (h *InviteHandler) Invite(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	// Authorize on the admin role. Per-request MFA is enforced at the Zitadel
	// login policy: the access token carries no amr claim (it lives only in the
	// ID token), so the gateway cannot re-verify a second factor here.
	if !domain.Can(p.Role, domain.PermEditUsers) {
		return echo.NewHTTPError(http.StatusForbidden, "requires admin")
	}

	var req InviteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if _, err := mail.ParseAddress(email); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email")
	}
	if !assignableRoles[role] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role")
	}

	ctx := c.Request().Context()
	userID, err := h.zit.CreateHumanUser(ctx, p.IDPOrgID, email, req.DisplayName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "create user")
	}
	if err := h.zit.GrantProjectRole(ctx, p.IDPOrgID, userID, role); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "grant role")
	}

	_ = h.audit.Write(ctx, audit.Entry{
		OrgID:        p.OrgID,
		ActorUserID:  p.UserID,
		Action:       "user.invited",
		ResourceType: "user",
		ResourceID:   userID,
		Metadata:     map[string]any{"email": email, "role": role},
	})

	return c.JSON(http.StatusCreated, InviteResponse{
		UserID: userID, Email: email, Role: role, Status: "invited",
	})
}

// Resend godoc
//
//	@Summary	Resend a pending user's invite email (admin + 2FA required)
//	@Tags		users
//	@Param		id	path	string	true	"Zitadel user id"
//	@Success	204
//	@Router		/users/{id}/resend [post]
func (h *InviteHandler) Resend(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	// Authorize on the admin role. Per-request MFA is enforced at the Zitadel
	// login policy: the access token carries no amr claim (it lives only in the
	// ID token), so the gateway cannot re-verify a second factor here.
	if !domain.Can(p.Role, domain.PermEditUsers) {
		return echo.NewHTTPError(http.StatusForbidden, "requires admin")
	}
	userID := c.Param("id")
	if userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing user id")
	}
	if err := h.zit.ResendInvite(c.Request().Context(), p.IDPOrgID, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "resend invite")
	}
	_ = h.audit.Write(c.Request().Context(), audit.Entry{
		OrgID:        p.OrgID,
		ActorUserID:  p.UserID,
		Action:       "user.invite_resent",
		ResourceType: "user",
		ResourceID:   userID,
	})
	return c.NoContent(http.StatusNoContent)
}
