// Package waves serves wave picking: batch several pick lists into one merged
// route, pick the summed quantity per SKU+zone once, then allocate picked units
// back across member orders (FIFO).
package waves

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/store"
)

// Store is the narrow store dependency the handler needs.
type Store interface {
	CreateWave(ctx context.Context, arg store.CreateWaveParams) (store.Wave, error)
	AssignListsToWave(ctx context.Context, arg store.AssignListsToWaveParams) error
	GetWave(ctx context.Context, arg store.GetWaveParams) (store.Wave, error)
	ListWavesByStore(ctx context.Context, arg store.ListWavesByStoreParams) ([]store.ListWavesByStoreRow, error)
	ListWaveMergedLines(ctx context.Context, arg store.ListWaveMergedLinesParams) ([]store.ListWaveMergedLinesRow, error)
	ListWaveStopMemberLines(ctx context.Context, arg store.ListWaveStopMemberLinesParams) ([]store.ListWaveStopMemberLinesRow, error)
	StartWave(ctx context.Context, arg store.StartWaveParams) (store.Wave, error)
	CompleteWave(ctx context.Context, arg store.CompleteWaveParams) (store.Wave, error)
	SetPickLinePicked(ctx context.Context, arg store.SetPickLinePickedParams) (store.PickListLine, error)
	AdjustItemLocationQty(ctx context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error)
	GetItemBySKU(ctx context.Context, arg store.GetItemBySKUParams) (store.Item, error)
	CountOpenTasksByTitle(ctx context.Context, arg store.CountOpenTasksByTitleParams) (int64, error)
	CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error)
}

// Auditor records an append-only audit entry. *audit.Writer satisfies it.
type Auditor interface {
	Write(ctx context.Context, e audit.Entry) error
}

// Handler serves wave-picking endpoints.
type Handler struct {
	q     Store
	pub   events.Publisher
	audit Auditor
}

// New creates a Handler. pub and audit may be nil to disable events/audit.
func New(q Store, pub events.Publisher, a Auditor) *Handler {
	return &Handler{q: q, pub: pub, audit: a}
}

// Register mounts wave routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.POST("/:storeID/waves", h.Create)
	g.GET("/:storeID/waves", h.List)
	g.GET("/:storeID/waves/:id", h.Board)
	g.POST("/:storeID/waves/:id/start", h.Start)
	g.PATCH("/:storeID/waves/:id/stops", h.PickStop)
	g.POST("/:storeID/waves/:id/complete", h.Complete)
}

// Stop is one merged stop on the wave route.
type Stop struct {
	Seq          int     `json:"seq"`
	SKU          string  `json:"sku"`
	ZoneID       *string `json:"zone_id,omitempty"`
	ZoneName     string  `json:"zone_name"`
	ZoneX        float64 `json:"zone_x"`
	ZoneY        float64 `json:"zone_y"`
	QtyRequested int32   `json:"qty_requested"`
	QtyPicked    int32   `json:"qty_picked"`
	OrderCount   int32   `json:"order_count"`
	Status       string  `json:"status"`
}

// Board is a wave with its merged, route-ordered stops.
type Board struct {
	ID        uuid.UUID `json:"id"`
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	Stops     []Stop    `json:"stops"`
}

// ListRow is one wave in the index.
type ListRow struct {
	ID        uuid.UUID `json:"id"`
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	ListCount int32     `json:"list_count"`
	CreatedAt string    `json:"created_at"`
}

func stopStatus(requested, picked int32) string {
	switch {
	case picked <= 0:
		return string(domain.PickLinePending)
	case picked >= requested:
		return string(domain.PickLinePicked)
	default:
		return string(domain.PickLineShort)
	}
}

// orderedStops resolves the wave's merged lines into a route-optimised slice.
func orderedStops(rows []store.ListWaveMergedLinesRow) []Stop {
	pstops := make([]domain.PickStop, 0, len(rows))
	byZoneSKU := make(map[string]store.ListWaveMergedLinesRow, len(rows))
	for _, r := range rows {
		key := r.Sku + "|" + uuid.UUID(r.ZoneID.Bytes).String()
		byZoneSKU[key] = r
		pstops = append(pstops, domain.PickStop{
			ZoneID: uuid.UUID(r.ZoneID.Bytes), SKU: r.Sku,
			QtyRequested: int(r.QtyRequested), X: r.ZoneX, Y: r.ZoneY,
		})
	}
	routed := domain.OptimizePickRoute(pstops, 0, 0)
	out := make([]Stop, 0, len(routed))
	for i, ps := range routed {
		key := ps.SKU + "|" + ps.ZoneID.String()
		r := byZoneSKU[key]
		zid := uuid.UUID(r.ZoneID.Bytes).String()
		out = append(out, Stop{
			Seq: i, SKU: r.Sku, ZoneID: &zid, ZoneName: r.ZoneName,
			ZoneX: r.ZoneX, ZoneY: r.ZoneY,
			QtyRequested: r.QtyRequested, QtyPicked: r.QtyPicked, OrderCount: r.OrderCount,
			Status: stopStatus(r.QtyRequested, r.QtyPicked),
		})
	}
	return out
}

type createRequest struct {
	Reference string   `json:"reference"`
	ListIDs   []string `json:"list_ids"`
}

// Create godoc
//
//	@Summary	Create a wave from pick lists and optimise its merged route
//	@Tags		waves
//	@Accept		json
//	@Produce	json
//	@Param		storeID	path		string			true	"Store UUID"
//	@Param		body	body		createRequest	true	"Reference + list ids"
//	@Success	201		{object}	Board
//	@Failure	400		{object}	map[string]string
//	@Router		/stores/{storeID}/waves [post]
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
	if len(req.ListIDs) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "a wave needs at least two pick lists")
	}
	ids := make([]uuid.UUID, 0, len(req.ListIDs))
	for _, s := range req.ListIDs {
		id, perr := uuid.Parse(s)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid list id")
		}
		ids = append(ids, id)
	}

	ctx := c.Request().Context()
	wave, err := h.q.CreateWave(ctx, store.CreateWaveParams{
		OrgID: p.OrgID, StoreID: storeID, Reference: req.Reference,
		CreatedBy: pgtype.UUID{Bytes: p.UserID, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create wave")
	}
	if err := h.q.AssignListsToWave(ctx, store.AssignListsToWaveParams{
		WaveID: pgtype.UUID{Bytes: wave.ID, Valid: true},
		OrgID:  p.OrgID, StoreID: storeID, ListIds: ids,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "assign lists")
	}
	return h.boardJSON(c, p.OrgID, wave, http.StatusCreated)
}

// List godoc
//
//	@Summary	List waves for a store
//	@Tags		waves
//	@Produce	json
//	@Param		storeID	path	string	true	"Store UUID"
//	@Success	200		{array}	ListRow
//	@Router		/stores/{storeID}/waves [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	rows, err := h.q.ListWavesByStore(c.Request().Context(), store.ListWavesByStoreParams{
		OrgID: p.OrgID, StoreID: storeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list waves")
	}
	out := make([]ListRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, ListRow{
			ID: r.ID, Reference: r.Reference, Status: r.Status,
			ListCount: r.ListCount, CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// Board godoc
//
//	@Summary	Get a wave with its merged route
//	@Tags		waves
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Wave UUID"
//	@Success	200		{object}	Board
//	@Failure	404		{object}	map[string]string
//	@Router		/stores/{storeID}/waves/{id} [get]
func (h *Handler) Board(c echo.Context) error {
	p, wave, err := h.load(c)
	if err != nil {
		return err
	}
	return h.boardJSON(c, p.OrgID, wave, http.StatusOK)
}

// Start godoc
//
//	@Summary	Mark a wave as in-progress
//	@Tags		waves
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Wave UUID"
//	@Success	200		{object}	map[string]string
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/waves/{id}/start [post]
func (h *Handler) Start(c echo.Context) error {
	p, wave, err := h.load(c)
	if err != nil {
		return err
	}
	done, serr := h.q.StartWave(c.Request().Context(), store.StartWaveParams{OrgID: p.OrgID, ID: wave.ID})
	if serr != nil {
		return echo.NewHTTPError(http.StatusConflict, "wave not open")
	}
	return c.JSON(http.StatusOK, map[string]string{"id": done.ID.String(), "status": done.Status})
}

type pickStopRequest struct {
	SKU       string `json:"sku"`
	ZoneID    string `json:"zone_id"`
	QtyPicked int    `json:"qty_picked"`
}

// PickStop godoc
//
//	@Summary	Confirm a merged stop: allocate picked units across member orders
//	@Tags		waves
//	@Accept		json
//	@Produce	json
//	@Param		storeID	path		string			true	"Store UUID"
//	@Param		id		path		string			true	"Wave UUID"
//	@Param		body	body		pickStopRequest	true	"SKU + zone + qty"
//	@Success	200		{object}	Board
//	@Failure	400		{object}	map[string]string
//	@Failure	404		{object}	map[string]string
//	@Router		/stores/{storeID}/waves/{id}/stops [patch]
func (h *Handler) PickStop(c echo.Context) error {
	p, wave, err := h.load(c)
	if err != nil {
		return err
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	var req pickStopRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.SKU == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sku required")
	}
	if req.QtyPicked < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "qty_picked must be non-negative")
	}
	zoneID, err := uuid.Parse(req.ZoneID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid zone_id")
	}

	ctx := c.Request().Context()
	members, err := h.q.ListWaveStopMemberLines(ctx, store.ListWaveStopMemberLinesParams{
		OrgID: p.OrgID, WaveID: pgtype.UUID{Bytes: wave.ID, Valid: true},
		Sku: req.SKU, ZoneID: pgtype.UUID{Bytes: zoneID, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list stop members")
	}
	if len(members) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "stop not found")
	}

	demand := 0
	lds := make([]domain.LineDemand, 0, len(members))
	for _, m := range members {
		demand += int(m.QtyRequested)
		lds = append(lds, domain.LineDemand{LineID: m.ID, Requested: int(m.QtyRequested)})
	}
	if req.QtyPicked > demand {
		return echo.NewHTTPError(http.StatusBadRequest, "qty_picked exceeds stop demand")
	}

	for _, fill := range domain.AllocatePick(req.QtyPicked, lds) {
		if _, err := h.q.SetPickLinePicked(ctx, store.SetPickLinePickedParams{
			QtyPicked: int32(fill.Picked), Status: string(fill.Status),
			OrgID: p.OrgID, ID: fill.LineID,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "allocate pick")
		}
	}

	// Decrement on-hand once for the whole merged pick.
	if req.QtyPicked > 0 {
		if _, err := h.q.AdjustItemLocationQty(ctx, store.AdjustItemLocationQtyParams{
			OrgID: p.OrgID, StoreID: storeID, ZoneID: zoneID, Sku: req.SKU, Qty: int32(-req.QtyPicked),
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "decrement inventory")
		}
		h.record(ctx, p, req.SKU, map[string]any{
			"wave_id": wave.ID.String(), "zone_id": zoneID.String(),
			"qty_picked": req.QtyPicked, "demand": demand,
		})
		if req.QtyPicked < demand {
			if err := h.maybeRestock(ctx, p.OrgID, storeID, zoneID, req.SKU); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "short-pick restock")
			}
		}
	}

	h.publishProgress(ctx, p.OrgID, storeID, wave.ID, req.SKU, req.QtyPicked, req.QtyPicked < demand)
	return h.boardJSON(c, p.OrgID, wave, http.StatusOK)
}

// Complete godoc
//
//	@Summary	Close a wave
//	@Tags		waves
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Wave UUID"
//	@Success	200		{object}	map[string]any
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/waves/{id}/complete [post]
func (h *Handler) Complete(c echo.Context) error {
	p, wave, err := h.load(c)
	if err != nil {
		return err
	}
	done, derr := h.q.CompleteWave(c.Request().Context(), store.CompleteWaveParams{OrgID: p.OrgID, ID: wave.ID})
	if derr != nil {
		return echo.NewHTTPError(http.StatusConflict, "wave already closed")
	}
	return c.JSON(http.StatusOK, map[string]any{"id": done.ID, "status": done.Status})
}

// boardJSON assembles and writes the merged-route board for a wave.
func (h *Handler) boardJSON(c echo.Context, orgID uuid.UUID, wave store.Wave, code int) error {
	rows, err := h.q.ListWaveMergedLines(c.Request().Context(), store.ListWaveMergedLinesParams{
		OrgID: orgID, WaveID: pgtype.UUID{Bytes: wave.ID, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "merged lines")
	}
	return c.JSON(code, Board{
		ID: wave.ID, Reference: wave.Reference, Status: wave.Status, Stops: orderedStops(rows),
	})
}

// load resolves the principal and wave, mapping a missing wave to 404.
func (h *Handler) load(c echo.Context) (*domain.Principal, store.Wave, error) {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return nil, store.Wave{}, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return nil, store.Wave{}, echo.NewHTTPError(http.StatusBadRequest, "invalid wave id")
	}
	wave, err := h.q.GetWave(c.Request().Context(), store.GetWaveParams{OrgID: p.OrgID, ID: id})
	if err != nil {
		return nil, store.Wave{}, echo.NewHTTPError(http.StatusNotFound, "wave not found")
	}
	return p, wave, nil
}

func (h *Handler) publishProgress(ctx context.Context, orgID, storeID, waveID uuid.UUID, sku string, qty int, short bool) {
	if h.pub == nil {
		return
	}
	status := string(domain.PickLinePicked)
	if short {
		status = string(domain.PickLineShort)
	}
	_ = h.pub.Publish(ctx, events.PickProgressSubject(orgID), events.PickProgress{
		OrgID: orgID, StoreID: storeID, ListID: waveID, SKU: sku,
		QtyPicked: qty, Status: status, TS: time.Now().UTC(),
	})
}

// maybeRestock raises a restock task when a short pick occurs and the SKU is known.
func (h *Handler) maybeRestock(ctx context.Context, orgID, storeID, zoneID uuid.UUID, sku string) error {
	if _, err := h.q.GetItemBySKU(ctx, store.GetItemBySKUParams{OrgID: orgID, Sku: sku}); err != nil {
		return nil
	}
	zoneArg := pgtype.UUID{Bytes: zoneID, Valid: true}
	title := domain.RestockTaskTitle(sku)
	n, err := h.q.CountOpenTasksByTitle(ctx, store.CountOpenTasksByTitleParams{
		OrgID: orgID, StoreID: storeID, ZoneID: zoneArg, Title: title,
	})
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = h.q.CreateTask(ctx, store.CreateTaskParams{
		OrgID: orgID, StoreID: storeID, ZoneID: zoneArg,
		Title: title, Status: string(domain.TaskStatusTodo), Priority: string(domain.TaskPriorityHigh),
	})
	return err
}

func (h *Handler) record(ctx context.Context, p *domain.Principal, sku string, meta map[string]any) {
	if h.audit == nil {
		return
	}
	_ = h.audit.Write(ctx, audit.Entry{
		OrgID:        p.OrgID,
		ActorUserID:  p.UserID,
		Action:       "inventory.wave_pick",
		ResourceType: "inventory",
		ResourceID:   sku,
		Metadata:     meta,
	})
}
