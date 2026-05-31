// Package sales serves read-only sales summary widgets for the dashboard.
package sales

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/store"
)

const sparkDays = 7

// Store is the narrow store dependency the handler needs.
type Store interface {
	SalesSummary(ctx context.Context, arg store.SalesSummaryParams) (store.SalesSummaryRow, error)
	SalesByDay(ctx context.Context, arg store.SalesByDayParams) ([]store.SalesByDayRow, error)
}

// Handler serves sales summary endpoints.
type Handler struct {
	q Store
}

// New creates a Handler.
func New(q Store) *Handler {
	return &Handler{q: q}
}

// Register mounts sales routes on the authenticated API group.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/sales/summary", h.Summary)
}

// SummaryResponse powers the dashboard sales widgets.
type SummaryResponse struct {
	RevenueCents int64   `json:"revenue_cents"`
	Units        int64   `json:"units"`
	Orders       int64   `json:"orders"`
	Spark        []int64 `json:"spark"`
}

// Summary godoc
//
//	@Summary	Today's sales totals plus a 7-day revenue sparkline
//	@Tags		sales
//	@Produce	json
//	@Success	200	{object}	SummaryResponse
//	@Router		/sales/summary [get]
func (h *Handler) Summary(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	ctx := c.Request().Context()
	now := time.Now().UTC()

	sum, err := h.q.SalesSummary(ctx, store.SalesSummaryParams{OrgID: p.OrgID, Ts: now.Add(-24 * time.Hour)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "sales summary")
	}

	since := now.AddDate(0, 0, -sparkDays).Truncate(24 * time.Hour)
	days, err := h.q.SalesByDay(ctx, store.SalesByDayParams{OrgID: p.OrgID, Ts: since})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "sales by day")
	}

	return c.JSON(http.StatusOK, SummaryResponse{
		RevenueCents: sum.RevenueCents,
		Units:        sum.Units,
		Orders:       sum.Orders,
		Spark:        spark(days, since, sparkDays),
	})
}

// spark builds a dense, zero-filled daily revenue series so the client renders a
// continuous sparkline even on days with no sales.
func spark(rows []store.SalesByDayRow, since time.Time, days int) []int64 {
	out := make([]int64, days)
	base := since.Truncate(24 * time.Hour)
	for _, r := range rows {
		d, ok := r.Day.(time.Time)
		if !ok {
			continue
		}
		idx := int(d.UTC().Truncate(24*time.Hour).Sub(base).Hours() / 24)
		if idx >= 0 && idx < days {
			out[idx] = r.RevenueCents
		}
	}
	return out
}
