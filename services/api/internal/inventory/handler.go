// Package inventory provides read and write endpoints for current on-hand stock.
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

// Store is the narrow store dependency the handler needs.
type Store interface {
	ListInventoryByStore(ctx context.Context, arg store.ListInventoryByStoreParams) ([]store.ListInventoryByStoreRow, error)
	UpsertItem(ctx context.Context, arg store.UpsertItemParams) (store.Item, error)
	AdjustItemLocationQty(ctx context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error)
}

// Handler serves inventory endpoints.
type Handler struct {
	q Store
}

// New creates a Handler.
func New(q Store) *Handler {
	return &Handler{q: q}
}

// Register mounts inventory routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.GET("/:storeID/inventory", h.List)
	g.POST("/:storeID/inventory", h.Add)
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

// AddRequest is the POST /stores/:storeID/inventory request body.
type AddRequest struct {
	ZoneID   string `json:"zone_id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Status   string `json:"status"`
	Qty      int32  `json:"qty"`
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

// Add godoc
//
//	@Summary		Add or adjust item quantity in a zone
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string		true	"Store UUID"
//	@Param			body	body		AddRequest	true	"Item + location"
//	@Success		201		{object}	Row
//	@Failure		400		{object}	map[string]string
//	@Router			/stores/{storeID}/inventory [post]
func (h *Handler) Add(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}

	var req AddRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	zoneID, err := uuid.Parse(req.ZoneID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid zone_id")
	}
	if req.SKU == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sku required")
	}
	if req.Qty <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "qty must be positive")
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	// Ensure item master exists.
	_, err = h.q.UpsertItem(c.Request().Context(), store.UpsertItemParams{
		OrgID:    p.OrgID,
		Sku:      req.SKU,
		Name:     req.Name,
		Category: req.Category,
		Status:   status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "upsert item")
	}

	// Adjust location qty (upsert, additive delta).
	loc, err := h.q.AdjustItemLocationQty(c.Request().Context(), store.AdjustItemLocationQtyParams{
		OrgID:   p.OrgID,
		StoreID: storeID,
		ZoneID:  zoneID,
		Sku:     req.SKU,
		Qty:     req.Qty,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "adjust qty")
	}

	return c.JSON(http.StatusCreated, Row{
		ID:        loc.ID,
		ZoneID:    loc.ZoneID,
		SKU:       loc.Sku,
		Name:      req.Name,
		Category:  req.Category,
		Status:    status,
		Qty:       loc.Qty,
		UpdatedAt: loc.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Velocity:  "cold",
	})
}
