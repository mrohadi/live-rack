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
	SetUserMFA(ctx context.Context, userID, orgID uuid.UUID, enabled bool) error
	TouchLastSeen(ctx context.Context, userID, orgID uuid.UUID) error
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
	g.GET("/users/members", h.Members)
	g.GET("/me", h.Me)
	g.POST("/me/2fa", h.SyncMFA)
}

// SyncMFARequest reports whether the caller completed a second factor. The amr
// claim lives only in the ID token, so the SPA syncs it here for roster + 2FA
// coverage. It is not a security gate — that stays at the Zitadel login policy.
type SyncMFARequest struct {
	Enabled bool `json:"enabled"`
}

// SyncMFA godoc
//
//	@Summary	Sync the caller's 2FA state from the ID-token amr claim
//	@Tags		users
//	@Accept		json
//	@Param		body	body	SyncMFARequest	true	"MFA state"
//	@Success	204
//	@Router		/me/2fa [post]
func (h *Handler) SyncMFA(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	var req SyncMFARequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.q.SetUserMFA(c.Request().Context(), p.UserID, p.OrgID, req.Enabled); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "sync mfa")
	}
	return c.NoContent(http.StatusNoContent)
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
	// Best-effort presence stamp; /me is polled on load.
	_ = h.q.TouchLastSeen(c.Request().Context(), p.UserID, p.OrgID)
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
	// Roster is admin-only — it exposes every member's role, presence, and 2FA state.
	if !domain.Can(p.Role, domain.PermEditUsers) {
		return echo.NewHTTPError(http.StatusForbidden, "requires admin")
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

// MemberRow is a minimal user record for assignee pickers — safe for all roles.
type MemberRow struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
}

// Members godoc
//
//	@Summary	Minimal member list for assignee pickers — accessible to all roles
//	@Tags		users
//	@Produce	json
//	@Success	200	{array}	MemberRow
//	@Router		/users/members [get]
func (h *Handler) Members(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	rows, err := h.q.ListUsersByOrg(c.Request().Context(), p.OrgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list members")
	}
	out := make([]MemberRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, MemberRow{
			ID:          r.ID.String(),
			DisplayName: r.DisplayName,
			Email:       r.Email,
			AvatarURL:   r.AvatarURL,
		})
	}
	return c.JSON(http.StatusOK, out)
}
