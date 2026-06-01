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

// StatsStore reads aggregate roster metrics.
type StatsStore interface {
	RosterStatsByOrg(ctx context.Context, orgID uuid.UUID) (store.RosterStats, error)
}

// PendingCounter reports the number of not-yet-verified invites for an org.
// *auth.ZitadelManagement satisfies it.
type PendingCounter interface {
	PendingInvites(ctx context.Context, idpOrgID string) (int, error)
}

// MetricsHandler serves the Users & Access header stat cards.
type MetricsHandler struct {
	q       StatsStore
	pending PendingCounter
}

// NewMetrics builds a MetricsHandler.
func NewMetrics(q StatsStore, pending PendingCounter) *MetricsHandler {
	return &MetricsHandler{q: q, pending: pending}
}

// Register mounts the stats route.
func (h *MetricsHandler) Register(g *echo.Group) {
	g.GET("/users/stats", h.Stats)
}

// StatsResponse backs the three header stat cards.
type StatsResponse struct {
	Members        int `json:"members"`
	Roles          int `json:"roles"`
	ActiveNow      int `json:"active_now"`
	PendingInvites int `json:"pending_invites"`
	// TwoFACoverage is a whole-number percentage of members with a second factor.
	TwoFACoverage int `json:"twofa_coverage"`
}

// Stats godoc
//
//	@Summary	Users & Access header metrics
//	@Tags		users
//	@Produce	json
//	@Success	200	{object}	StatsResponse
//	@Router		/users/stats [get]
func (h *MetricsHandler) Stats(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.Can(p.Role, domain.PermViewDashboards) {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient role")
	}

	s, err := h.q.RosterStatsByOrg(c.Request().Context(), p.OrgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "roster stats")
	}

	coverage := 0
	if s.Members > 0 {
		coverage = s.MFAUsers * 100 / s.Members
	}

	// Pending invites come from local pending rows (written at invite time). Fall
	// back to Zitadel's count only when we have no local pending rows.
	pending := s.Pending
	if pending == 0 && h.pending != nil && p.IDPOrgID != "" {
		if n, err := h.pending.PendingInvites(c.Request().Context(), p.IDPOrgID); err == nil {
			pending = n
		}
	}

	return c.JSON(http.StatusOK, StatsResponse{
		Members:        s.Members,
		Roles:          s.Roles,
		ActiveNow:      s.ActiveNow,
		PendingInvites: pending,
		TwoFACoverage:  coverage,
	})
}
