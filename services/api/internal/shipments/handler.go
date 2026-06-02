// Package shipments serves the packing + dispatch stage: turn a completed pick
// list into a shipment, record carrier + tracking, and dispatch it.
package shipments

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// Store is the narrow store dependency the handler needs.
type Store interface {
	GetPickList(ctx context.Context, arg store.GetPickListParams) (store.PickList, error)
	ListPickListLines(ctx context.Context, listID uuid.UUID) ([]store.ListPickListLinesRow, error)
	CreateShipment(ctx context.Context, arg store.CreateShipmentParams) (store.Shipment, error)
	AddShipmentItem(ctx context.Context, arg store.AddShipmentItemParams) (store.ShipmentItem, error)
	GetShipment(ctx context.Context, arg store.GetShipmentParams) (store.Shipment, error)
	ListShipmentsByStore(ctx context.Context, arg store.ListShipmentsByStoreParams) ([]store.ListShipmentsByStoreRow, error)
	ListShipmentItems(ctx context.Context, shipmentID uuid.UUID) ([]store.ListShipmentItemsRow, error)
	MarkShipmentPacked(ctx context.Context, arg store.MarkShipmentPackedParams) (store.Shipment, error)
	MarkShipmentDispatched(ctx context.Context, arg store.MarkShipmentDispatchedParams) (store.Shipment, error)
	CancelShipment(ctx context.Context, arg store.CancelShipmentParams) (store.Shipment, error)
}

// Auditor records an append-only audit entry. *audit.Writer satisfies it.
type Auditor interface {
	Write(ctx context.Context, e audit.Entry) error
}

// Handler serves shipment endpoints.
type Handler struct {
	q     Store
	audit Auditor
}

// New creates a Handler. audit may be nil to disable audit logging.
func New(q Store, a Auditor) *Handler {
	return &Handler{q: q, audit: a}
}

// Register mounts shipment routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.POST("/:storeID/shipments", h.Create)
	g.GET("/:storeID/shipments", h.List)
	g.GET("/:storeID/shipments/:id", h.Board)
	g.POST("/:storeID/shipments/:id/pack", h.Pack)
	g.POST("/:storeID/shipments/:id/dispatch", h.Dispatch)
	g.POST("/:storeID/shipments/:id/cancel", h.Cancel)
}

// Item is one packed line.
type Item struct {
	SKU string `json:"sku"`
	Qty int32  `json:"qty"`
}

// Board is a shipment with its packed items.
type Board struct {
	ID             uuid.UUID `json:"id"`
	Reference      string    `json:"reference"`
	Status         string    `json:"status"`
	Carrier        string    `json:"carrier"`
	TrackingNumber string    `json:"tracking_number"`
	Items          []Item    `json:"items"`
}

// ListRow is one shipment in the index.
type ListRow struct {
	ID             uuid.UUID `json:"id"`
	Reference      string    `json:"reference"`
	Status         string    `json:"status"`
	Carrier        string    `json:"carrier"`
	TrackingNumber string    `json:"tracking_number"`
	ItemCount      int32     `json:"item_count"`
	CreatedAt      string    `json:"created_at"`
}

type createRequest struct {
	PickListID string `json:"pick_list_id"`
	Reference  string `json:"reference"`
}

// Create godoc
//
//	@Summary	Create a shipment from a completed pick list (snapshots picked lines)
//	@Tags		shipments
//	@Accept		json
//	@Produce	json
//	@Param		storeID	path		string			true	"Store UUID"
//	@Param		body	body		createRequest	true	"Pick list + reference"
//	@Success	201		{object}	Board
//	@Failure	400		{object}	map[string]string
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/shipments [post]
func (h *Handler) Create(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	var req createRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	listID, err := uuid.Parse(req.PickListID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pick_list_id")
	}

	ctx := c.Request().Context()
	list, err := h.q.GetPickList(ctx, store.GetPickListParams{OrgID: p.OrgID, ID: listID})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "pick list not found")
	}
	if list.Status != string(domain.PickListCompleted) {
		return echo.NewHTTPError(http.StatusConflict, "pick list not completed")
	}
	lines, err := h.q.ListPickListLines(ctx, list.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}

	ship, err := h.q.CreateShipment(ctx, store.CreateShipmentParams{
		OrgID: p.OrgID, StoreID: storeID,
		PickListID: pgtype.UUID{Bytes: list.ID, Valid: true},
		Reference:  req.Reference,
		CreatedBy:  pgtype.UUID{Bytes: p.UserID, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create shipment")
	}
	packed := 0
	for _, l := range lines {
		if l.QtyPicked <= 0 {
			continue
		}
		if _, err := h.q.AddShipmentItem(ctx, store.AddShipmentItemParams{
			ShipmentID: ship.ID, OrgID: p.OrgID, Sku: l.Sku, Qty: l.QtyPicked,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "add item")
		}
		packed++
	}
	if packed == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "pick list has no picked items")
	}
	h.record(ctx, p, ship.ID.String(), "inventory.shipment_create", map[string]any{
		"pick_list_id": list.ID.String(), "items": packed,
	})
	return h.boardJSON(c, ship, http.StatusCreated)
}

// List godoc
//
//	@Summary	List shipments for a store
//	@Tags		shipments
//	@Produce	json
//	@Param		storeID	path	string	true	"Store UUID"
//	@Success	200		{array}	ListRow
//	@Router		/stores/{storeID}/shipments [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	rows, err := h.q.ListShipmentsByStore(c.Request().Context(), store.ListShipmentsByStoreParams{
		OrgID: p.OrgID, StoreID: storeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list shipments")
	}
	out := make([]ListRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, ListRow{
			ID: r.ID, Reference: r.Reference, Status: r.Status,
			Carrier: r.Carrier, TrackingNumber: r.TrackingNumber,
			ItemCount: r.ItemCount, CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// Board godoc
//
//	@Summary	Get a shipment with its packed items
//	@Tags		shipments
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Shipment UUID"
//	@Success	200		{object}	Board
//	@Failure	404		{object}	map[string]string
//	@Router		/stores/{storeID}/shipments/{id} [get]
func (h *Handler) Board(c echo.Context) error {
	_, ship, err := h.load(c)
	if err != nil {
		return err
	}
	return h.boardJSON(c, ship, http.StatusOK)
}

// Pack godoc
//
//	@Summary	Mark a shipment packed (ready to dispatch)
//	@Tags		shipments
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Shipment UUID"
//	@Success	200		{object}	map[string]string
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/shipments/{id}/pack [post]
func (h *Handler) Pack(c echo.Context) error {
	p, ship, err := h.load(c)
	if err != nil {
		return err
	}
	done, derr := h.q.MarkShipmentPacked(c.Request().Context(), store.MarkShipmentPackedParams{OrgID: p.OrgID, ID: ship.ID})
	if derr != nil {
		return echo.NewHTTPError(http.StatusConflict, "shipment not in packing state")
	}
	return c.JSON(http.StatusOK, map[string]string{"id": done.ID.String(), "status": done.Status})
}

type dispatchRequest struct {
	Carrier        string `json:"carrier"`
	TrackingNumber string `json:"tracking_number"`
}

// Dispatch godoc
//
//	@Summary	Dispatch a packed shipment with carrier + tracking
//	@Tags		shipments
//	@Accept		json
//	@Produce	json
//	@Param		storeID	path		string			true	"Store UUID"
//	@Param		id		path		string			true	"Shipment UUID"
//	@Param		body	body		dispatchRequest	true	"Carrier + tracking"
//	@Success	200		{object}	Board
//	@Failure	400		{object}	map[string]string
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/shipments/{id}/dispatch [post]
func (h *Handler) Dispatch(c echo.Context) error {
	p, ship, err := h.load(c)
	if err != nil {
		return err
	}
	if !domain.CanDispatch(domain.ShipmentStatus(ship.Status)) {
		return echo.NewHTTPError(http.StatusConflict, "shipment must be packed first")
	}
	var req dispatchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Carrier == "" || req.TrackingNumber == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "carrier and tracking_number required")
	}

	done, derr := h.q.MarkShipmentDispatched(c.Request().Context(), store.MarkShipmentDispatchedParams{
		Carrier: req.Carrier, TrackingNumber: req.TrackingNumber, OrgID: p.OrgID, ID: ship.ID,
	})
	if derr != nil {
		if errors.Is(derr, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusConflict, "shipment not packed")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "dispatch")
	}
	h.record(c.Request().Context(), p, ship.ID.String(), "inventory.shipment_dispatch", map[string]any{
		"carrier": req.Carrier, "tracking_number": req.TrackingNumber,
	})
	return h.boardJSON(c, done, http.StatusOK)
}

// Cancel godoc
//
//	@Summary	Cancel a shipment before it ships
//	@Tags		shipments
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Shipment UUID"
//	@Success	200		{object}	map[string]string
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/shipments/{id}/cancel [post]
func (h *Handler) Cancel(c echo.Context) error {
	p, ship, err := h.load(c)
	if err != nil {
		return err
	}
	done, derr := h.q.CancelShipment(c.Request().Context(), store.CancelShipmentParams{OrgID: p.OrgID, ID: ship.ID})
	if derr != nil {
		return echo.NewHTTPError(http.StatusConflict, "shipment already shipped or cancelled")
	}
	return c.JSON(http.StatusOK, map[string]string{"id": done.ID.String(), "status": done.Status})
}

// boardJSON assembles and writes a shipment board.
func (h *Handler) boardJSON(c echo.Context, ship store.Shipment, code int) error {
	items, err := h.q.ListShipmentItems(c.Request().Context(), ship.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list items")
	}
	rows := make([]Item, 0, len(items))
	for _, it := range items {
		rows = append(rows, Item{SKU: it.Sku, Qty: it.Qty})
	}
	return c.JSON(code, Board{
		ID: ship.ID, Reference: ship.Reference, Status: ship.Status,
		Carrier: ship.Carrier, TrackingNumber: ship.TrackingNumber, Items: rows,
	})
}

// load resolves the principal and shipment, mapping a missing shipment to 404.
func (h *Handler) load(c echo.Context) (*domain.Principal, store.Shipment, error) {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return nil, store.Shipment{}, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return nil, store.Shipment{}, echo.NewHTTPError(http.StatusBadRequest, "invalid shipment id")
	}
	ship, err := h.q.GetShipment(c.Request().Context(), store.GetShipmentParams{OrgID: p.OrgID, ID: id})
	if err != nil {
		return nil, store.Shipment{}, echo.NewHTTPError(http.StatusNotFound, "shipment not found")
	}
	return p, ship, nil
}

func (h *Handler) record(ctx context.Context, p *domain.Principal, id, action string, meta map[string]any) {
	if h.audit == nil {
		return
	}
	_ = h.audit.Write(ctx, audit.Entry{
		OrgID:        p.OrgID,
		ActorUserID:  p.UserID,
		Action:       action,
		ResourceType: "shipment",
		ResourceID:   id,
		Metadata:     meta,
	})
}
