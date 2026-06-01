package users

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// AccessStore is the persistence the access actions need.
type AccessStore interface {
	ListAudit(ctx context.Context, orgID uuid.UUID, actor *uuid.UUID, limit int) ([]store.AuditRow, error)
	SetUserRole(ctx context.Context, orgID, userID uuid.UUID, role string) error
}

// AccessZit is the Zitadel surface the access actions need.
type AccessZit interface {
	GrantProjectRole(ctx context.Context, orgID, userID, role string) error
	SendPasswordReset(ctx context.Context, orgID, userID string) error
}

// AccessHandler serves the audit trail and member access actions.
type AccessHandler struct {
	q     AccessStore
	zit   AccessZit
	audit Auditor
}

// NewAccess builds an AccessHandler.
func NewAccess(q AccessStore, zit AccessZit, a Auditor) *AccessHandler {
	return &AccessHandler{q: q, zit: zit, audit: a}
}

// Register mounts the audit + access-action routes.
func (h *AccessHandler) Register(g *echo.Group) {
	g.GET("/audit", h.Audit)
	g.PATCH("/users/:id/role", h.SetRole)
	g.POST("/users/:id/reset-password", h.ResetPassword)
}

// Audit godoc
//
//	@Summary	Recent audit-trail entries (optionally filtered by actor)
//	@Tags		users
//	@Produce	json
//	@Param		actor	query	string	false	"Actor user id"
//	@Param		limit	query	int		false	"Max rows (default 20)"
//	@Success	200	{array}	store.AuditRow
//	@Router		/audit [get]
func (h *AccessHandler) Audit(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.Can(p.Role, domain.PermViewDashboards) {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient role")
	}

	var actor *uuid.UUID
	if a := c.QueryParam("actor"); a != "" {
		id, err := uuid.Parse(a)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid actor")
		}
		actor = &id
	}
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	rows, err := h.q.ListAudit(c.Request().Context(), p.OrgID, actor, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list audit")
	}
	if rows == nil {
		rows = []store.AuditRow{}
	}
	return c.JSON(http.StatusOK, rows)
}

// SetRoleRequest changes a member's role. idp_user_id re-grants the role in
// Zitadel so the next token carries it.
type SetRoleRequest struct {
	Role      string `json:"role"`
	IDPUserID string `json:"idp_user_id"`
}

// SetRole godoc
//
//	@Summary	Change a member's role (admin required)
//	@Tags		users
//	@Accept		json
//	@Param		id		path	string			true	"Internal user id"
//	@Param		body	body	SetRoleRequest	true	"New role"
//	@Success	204
//	@Router		/users/{id}/role [patch]
func (h *AccessHandler) SetRole(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.Can(p.Role, domain.PermEditUsers) {
		return echo.NewHTTPError(http.StatusForbidden, "requires admin")
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	var req SetRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if !assignableRoles[req.Role] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role")
	}

	ctx := c.Request().Context()
	if err := h.q.SetUserRole(ctx, p.OrgID, userID, req.Role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "set role")
	}
	// Re-grant in Zitadel so the next token reflects the change (best-effort).
	if req.IDPUserID != "" {
		_ = h.zit.GrantProjectRole(ctx, p.IDPOrgID, req.IDPUserID, req.Role)
	}
	_ = h.audit.Write(ctx, audit.Entry{
		OrgID: p.OrgID, ActorUserID: p.UserID, Action: "user.role_changed",
		ResourceType: "user", ResourceID: userID.String(),
		Metadata: map[string]any{"role": req.Role},
	})
	return c.NoContent(http.StatusNoContent)
}

// ResetPassword godoc
//
//	@Summary	Email a member a password-reset link (admin required)
//	@Tags		users
//	@Param		id	path	string	true	"Zitadel user id"
//	@Success	204
//	@Router		/users/{id}/reset-password [post]
func (h *AccessHandler) ResetPassword(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.Can(p.Role, domain.PermEditUsers) {
		return echo.NewHTTPError(http.StatusForbidden, "requires admin")
	}
	idpUserID := c.Param("id")
	if idpUserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing user id")
	}
	if err := h.zit.SendPasswordReset(c.Request().Context(), p.IDPOrgID, idpUserID); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "reset password")
	}
	_ = h.audit.Write(c.Request().Context(), audit.Entry{
		OrgID: p.OrgID, ActorUserID: p.UserID, Action: "user.password_reset",
		ResourceType: "user", ResourceID: idpUserID,
	})
	return c.NoContent(http.StatusNoContent)
}
