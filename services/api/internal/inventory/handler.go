// Package inventory provides read endpoints for current on-hand stock.
package inventory

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// Lister is the narrow store dependency the handler needs.
type Lister interface {
	ListInventoryByStore(ctx context.Context, arg store.ListInventoryByStoreParams) ([]store.ListInventoryByStoreRow, error)
}

// Handler serves inventory read endpoints.
type Handler struct {
	q Lister
}

// New creates a Handler.
func New(q Lister) *Handler {
	return &Handler{q: q}
}

// Register mounts inventory routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.GET("/:storeID/inventory", h.List)
}

// Row is one on-hand line returned to the client.
type Row struct {
	ID        uuid.UUID `json:"id"`
	ZoneID    uuid.UUID `json:"zone_id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Status    string    `json:"status"`
	Qty       int32     `json:"qty"`
	UpdatedAt string    `json:"updated_at"`
	Velocity  string    `json:"velocity"`
}

// List godoc
//
//	@Summary		List current on-hand inventory for a store
//	@Tags			inventory
//	@Produce		json
//	@Param			storeID	path		string	true	"Store UUID"
//	@Success		200		{array}		Row
//	@Failure		400		{object}	map[string]string
//	@Router			/stores/{storeID}/inventory [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}

	rows, err := h.q.ListInventoryByStore(c.Request().Context(), store.ListInventoryByStoreParams{
		OrgID:   p.OrgID,
		StoreID: storeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list inventory")
	}

	out := make([]Row, 0, len(rows))
	for _, r := range rows {
		out = append(out, Row{
			ID:        r.ID,
			ZoneID:    r.ZoneID,
			SKU:       r.Sku,
			Name:      r.Name,
			Category:  r.Category,
			Status:    r.Status,
			Qty:       r.Qty,
			UpdatedAt: r.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			Velocity:  string(domain.VelocityFromPicks(int(r.Picks7d), int(r.Picks30d))),
		})
	}
	return c.JSON(http.StatusOK, out)
}
