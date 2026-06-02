// Package inventory provides read and write endpoints for current on-hand stock.
package inventory

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	DecrementItemLocationQty(ctx context.Context, arg store.DecrementItemLocationQtyParams) (store.ItemLocation, error)
	GetItemBySKU(ctx context.Context, arg store.GetItemBySKUParams) (store.Item, error)
	ListItemLocationsBySKU(ctx context.Context, arg store.ListItemLocationsBySKUParams) ([]store.ListItemLocationsBySKURow, error)
	ListScanEventsBySKU(ctx context.Context, arg store.ListScanEventsBySKUParams) ([]store.ScanEvent, error)
	UpdateItem(ctx context.Context, arg store.UpdateItemParams) (store.Item, error)
	SetItemLocationQty(ctx context.Context, arg store.SetItemLocationQtyParams) (store.ItemLocation, error)
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
	g.POST("/:storeID/inventory/transfer", h.Transfer)
	g.GET("/:storeID/inventory/:sku", h.Detail)
	g.PATCH("/:storeID/inventory/:sku", h.EditItem)
	g.PATCH("/:storeID/inventory/:sku/qty", h.AdjustQty)
}

// Row is one on-hand line returned to the client.
type Row struct {
	ID           uuid.UUID `json:"id"`
	ZoneID       uuid.UUID `json:"zone_id"`
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Status       string    `json:"status"`
	Qty          int32     `json:"qty"`
	ReorderPoint int32     `json:"reorder_point"`
	StockStatus  string    `json:"stock_status"`
	UpdatedAt    string    `json:"updated_at"`
	Velocity     string    `json:"velocity"`
}

// AddRequest is the POST /stores/:storeID/inventory request body.
type AddRequest struct {
	ZoneID       string `json:"zone_id"`
	SKU          string `json:"sku"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	Qty          int32  `json:"qty"`
	ReorderPoint int32  `json:"reorder_point"`
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
			ID:           r.ID,
			ZoneID:       r.ZoneID,
			SKU:          r.Sku,
			Name:         r.Name,
			Category:     r.Category,
			Status:       r.Status,
			Qty:          r.Qty,
			ReorderPoint: r.ReorderPoint,
			StockStatus:  string(domain.StockStatusFromQty(int(r.Qty), int(r.ReorderPoint))),
			UpdatedAt:    r.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			Velocity:     string(domain.VelocityFromPicks(int(r.Picks7d), int(r.Picks30d))),
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

	if req.ReorderPoint < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "reorder_point must be non-negative")
	}

	// Ensure item master exists.
	_, err = h.q.UpsertItem(c.Request().Context(), store.UpsertItemParams{
		OrgID:        p.OrgID,
		Sku:          req.SKU,
		Name:         req.Name,
		Category:     req.Category,
		Status:       status,
		ReorderPoint: req.ReorderPoint,
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
		ID:           loc.ID,
		ZoneID:       loc.ZoneID,
		SKU:          loc.Sku,
		Name:         req.Name,
		Category:     req.Category,
		Status:       status,
		Qty:          loc.Qty,
		ReorderPoint: req.ReorderPoint,
		StockStatus:  string(domain.StockStatusFromQty(int(loc.Qty), int(req.ReorderPoint))),
		UpdatedAt:    loc.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Velocity:     "cold",
	})
}

// TransferRequest is the POST /stores/:storeID/inventory/transfer body.
type TransferRequest struct {
	SKU        string `json:"sku"`
	FromZoneID string `json:"from_zone_id"`
	ToZoneID   string `json:"to_zone_id"`
	Qty        int32  `json:"qty"`
}

// TransferResponse confirms a completed move.
type TransferResponse struct {
	SKU        string    `json:"sku"`
	FromZoneID uuid.UUID `json:"from_zone_id"`
	ToZoneID   uuid.UUID `json:"to_zone_id"`
	Qty        int32     `json:"qty"`
}

// Transfer godoc
//
//	@Summary		Move stock of a SKU from one zone to another
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Param			body	body		TransferRequest	true	"Transfer"
//	@Success		200		{object}	TransferResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Router			/stores/{storeID}/inventory/transfer [post]
func (h *Handler) Transfer(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}

	var req TransferRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.SKU == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sku required")
	}
	if req.Qty <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "qty must be positive")
	}

	fromZone, err := uuid.Parse(req.FromZoneID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid from_zone_id")
	}
	toZone, err := uuid.Parse(req.ToZoneID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid to_zone_id")
	}
	if fromZone == toZone {
		return echo.NewHTTPError(http.StatusBadRequest, "source and destination zones must differ")
	}

	ctx := c.Request().Context()

	// Guarded source decrement: fails (no rows) when stock is insufficient.
	if _, err := h.q.DecrementItemLocationQty(ctx, store.DecrementItemLocationQtyParams{
		OrgID:  p.OrgID,
		ZoneID: fromZone,
		Sku:    req.SKU,
		Qty:    req.Qty,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusConflict, "insufficient stock in source zone")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "decrement source")
	}

	// Credit the destination zone.
	if _, err := h.q.AdjustItemLocationQty(ctx, store.AdjustItemLocationQtyParams{
		OrgID:   p.OrgID,
		StoreID: storeID,
		ZoneID:  toZone,
		Sku:     req.SKU,
		Qty:     req.Qty,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "credit destination")
	}

	return c.JSON(http.StatusOK, TransferResponse{
		SKU:        req.SKU,
		FromZoneID: fromZone,
		ToZoneID:   toZone,
		Qty:        req.Qty,
	})
}

const detailScanLimit = 20

// LocationRow is one zone's on-hand line in the item detail view.
type LocationRow struct {
	ZoneID      uuid.UUID `json:"zone_id"`
	ZoneName    string    `json:"zone_name"`
	Qty         int32     `json:"qty"`
	StockStatus string    `json:"stock_status"`
	UpdatedAt   string    `json:"updated_at"`
}

// ScanRow is one scan-timeline entry in the item detail view.
type ScanRow struct {
	TS        string `json:"ts"`
	ZoneID    string `json:"zone_id"`
	ScannerID string `json:"scanner_id"`
	Action    string `json:"action"`
	Valid     bool   `json:"valid"`
	Reason    string `json:"reason,omitempty"`
}

// DetailResponse is the full item drawer payload.
type DetailResponse struct {
	SKU          string        `json:"sku"`
	Name         string        `json:"name"`
	Category     string        `json:"category"`
	Status       string        `json:"status"`
	ReorderPoint int32         `json:"reorder_point"`
	TotalQty     int32         `json:"total_qty"`
	StockStatus  string        `json:"stock_status"`
	Locations    []LocationRow `json:"locations"`
	RecentScans  []ScanRow     `json:"recent_scans"`
}

// Detail godoc
//
//	@Summary		Item detail — per-zone on-hand + recent scan timeline
//	@Tags			inventory
//	@Produce		json
//	@Param			storeID	path		string	true	"Store UUID"
//	@Param			sku		path		string	true	"SKU"
//	@Success		200		{object}	DetailResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/stores/{storeID}/inventory/{sku} [get]
func (h *Handler) Detail(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	sku := c.Param("sku")
	if sku == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sku required")
	}

	ctx := c.Request().Context()

	item, err := h.q.GetItemBySKU(ctx, store.GetItemBySKUParams{OrgID: p.OrgID, Sku: sku})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "item not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "get item")
	}

	locs, err := h.q.ListItemLocationsBySKU(ctx, store.ListItemLocationsBySKUParams{
		OrgID:   p.OrgID,
		StoreID: storeID,
		Sku:     sku,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list locations")
	}

	scans, err := h.q.ListScanEventsBySKU(ctx, store.ListScanEventsBySKUParams{
		OrgID:   p.OrgID,
		StoreID: storeID,
		Sku:     sku,
		Limit:   detailScanLimit,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list scans")
	}

	rp := int(item.ReorderPoint)
	var total int32
	locations := make([]LocationRow, 0, len(locs))
	for _, l := range locs {
		total += l.Qty
		locations = append(locations, LocationRow{
			ZoneID:      l.ZoneID,
			ZoneName:    l.ZoneName,
			Qty:         l.Qty,
			StockStatus: string(domain.StockStatusFromQty(int(l.Qty), rp)),
			UpdatedAt:   l.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	recent := make([]ScanRow, 0, len(scans))
	for _, s := range scans {
		recent = append(recent, ScanRow{
			TS:        s.Ts.UTC().Format("2006-01-02T15:04:05Z07:00"),
			ZoneID:    s.ZoneID.String(),
			ScannerID: s.ScannerID,
			Action:    s.Action,
			Valid:     s.Valid,
			Reason:    s.Reason,
		})
	}

	return c.JSON(http.StatusOK, DetailResponse{
		SKU:          item.Sku,
		Name:         item.Name,
		Category:     item.Category,
		Status:       item.Status,
		ReorderPoint: item.ReorderPoint,
		TotalQty:     total,
		StockStatus:  string(domain.StockStatusFromQty(int(total), rp)),
		Locations:    locations,
		RecentScans:  recent,
	})
}

// EditItemRequest is the PATCH /inventory/:sku body.
type EditItemRequest struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	ReorderPoint int32  `json:"reorder_point"`
}

// EditItem godoc
//
//	@Summary		Edit master-catalog fields for a SKU
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Param			sku		path		string			true	"SKU"
//	@Param			body	body		EditItemRequest	true	"Item fields"
//	@Success		200		{object}	map[string]any
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/stores/{storeID}/inventory/{sku} [patch]
func (h *Handler) EditItem(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	sku := c.Param("sku")
	if sku == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sku required")
	}

	var req EditItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if !domain.ItemStatus(req.Status).Valid() {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}
	if req.ReorderPoint < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "reorder_point must be non-negative")
	}

	item, err := h.q.UpdateItem(c.Request().Context(), store.UpdateItemParams{
		OrgID:        p.OrgID,
		Sku:          sku,
		Name:         req.Name,
		Category:     req.Category,
		Status:       req.Status,
		ReorderPoint: req.ReorderPoint,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "item not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "update item")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"sku":           item.Sku,
		"name":          item.Name,
		"category":      item.Category,
		"status":        item.Status,
		"reorder_point": item.ReorderPoint,
	})
}

// AdjustQtyRequest is the PATCH /inventory/:sku/qty body — absolute set.
type AdjustQtyRequest struct {
	ZoneID string `json:"zone_id"`
	Qty    int32  `json:"qty"`
}

// AdjustQty godoc
//
//	@Summary		Manually correct on-hand qty in a zone (shrinkage, damage, count)
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string				true	"Store UUID"
//	@Param			sku		path		string				true	"SKU"
//	@Param			body	body		AdjustQtyRequest	true	"Zone + absolute qty"
//	@Success		200		{object}	map[string]any
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/stores/{storeID}/inventory/{sku}/qty [patch]
func (h *Handler) AdjustQty(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	sku := c.Param("sku")
	if sku == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sku required")
	}

	var req AdjustQtyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Qty < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "qty must be non-negative")
	}
	zoneID, err := uuid.Parse(req.ZoneID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid zone_id")
	}

	loc, err := h.q.SetItemLocationQty(c.Request().Context(), store.SetItemLocationQtyParams{
		OrgID:   p.OrgID,
		StoreID: storeID,
		ZoneID:  zoneID,
		Sku:     sku,
		Qty:     req.Qty,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "location not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "set qty")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"sku":     loc.Sku,
		"zone_id": loc.ZoneID,
		"qty":     loc.Qty,
	})
}
