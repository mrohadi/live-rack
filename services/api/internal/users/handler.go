// Package users serves the Users & Access screen: the org roster and the
// caller's own capabilities (role, permissions, MFA, scopes).
package users

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// Store is the narrow dependency the handler needs.
type Store interface {
	ListUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]store.UserListRow, error)
}

// Handler serves user + capability endpoints.
type Handler struct {
	q Store
}

// New creates a Handler.
func New(q Store) *Handler {
	return &Handler{q: q}
}

// Register mounts routes on the authenticated API group.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/users", h.List)
	g.GET("/me", h.Me)
}

// CapabilitiesResponse describes the caller's effective access.
type CapabilitiesResponse struct {
	UserID      string   `json:"user_id"`
	Role        string   `json:"role"`
	MFAVerified bool     `json:"mfa_verified"`
	Permissions []string `json:"permissions"`
	StoreScoped bool     `json:"store_scoped"`
	ZoneScoped  bool     `json:"zone_scoped"`
}

// Me godoc
//
//	@Summary	The caller's role, permissions, MFA state, and scopes
//	@Tags		users
//	@Produce	json
//	@Success	200	{object}	CapabilitiesResponse
//	@Router		/me [get]
func (h *Handler) Me(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	perms := domain.Permissions(p.Role)
	out := CapabilitiesResponse{
		UserID:      p.UserID.String(),
		Role:        string(p.Role),
		MFAVerified: p.MFAVerified,
		Permissions: make([]string, len(perms)),
		StoreScoped: len(p.StoreIDs) > 0,
		ZoneScoped:  len(p.ZoneIDs) > 0,
	}
	for i, perm := range perms {
		out.Permissions[i] = string(perm)
	}
	return c.JSON(http.StatusOK, out)
}

// List godoc
//
//	@Summary	Org user roster with bound roles
//	@Tags		users
//	@Produce	json
//	@Success	200	{array}	store.UserListRow
//	@Router		/users [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.Can(p.Role, domain.PermViewDashboards) {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient role")
	}
	rows, err := h.q.ListUsersByOrg(c.Request().Context(), p.OrgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list users")
	}
	if rows == nil {
		rows = []store.UserListRow{}
	}
	return c.JSON(http.StatusOK, rows)
}
