// Package picking serves pick-list fulfilment: build a map-optimised pick route,
// confirm picks by scan, decrement on-hand inventory, and flag short picks.
package picking

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
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/store"
)

// Store is the narrow store dependency the handler needs.
type Store interface {
	CreatePickList(ctx context.Context, arg store.CreatePickListParams) (store.PickList, error)
	AddPickLine(ctx context.Context, arg store.AddPickLineParams) (store.PickListLine, error)
	GetPickList(ctx context.Context, arg store.GetPickListParams) (store.PickList, error)
	ListPickListsByStore(ctx context.Context, arg store.ListPickListsByStoreParams) ([]store.ListPickListsByStoreRow, error)
	ListPickListLines(ctx context.Context, listID uuid.UUID) ([]store.ListPickListLinesRow, error)
	SetPickLinePicked(ctx context.Context, arg store.SetPickLinePickedParams) (store.PickListLine, error)
	StartPickList(ctx context.Context, arg store.StartPickListParams) (store.PickList, error)
	CompletePickList(ctx context.Context, arg store.CompletePickListParams) (store.PickList, error)
	ResolvePickSource(ctx context.Context, arg store.ResolvePickSourceParams) (store.ResolvePickSourceRow, error)
	AdjustItemLocationQty(ctx context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error)
	GetItemBySKU(ctx context.Context, arg store.GetItemBySKUParams) (store.Item, error)
	CountOpenTasksByTitle(ctx context.Context, arg store.CountOpenTasksByTitleParams) (int64, error)
	CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error)
}

// Auditor records an append-only audit entry. *audit.Writer satisfies it.
type Auditor interface {
	Write(ctx context.Context, e audit.Entry) error
}

// Handler serves pick-list endpoints.
type Handler struct {
	q     Store
	pub   events.Publisher
	audit Auditor
}

// New creates a Handler. pub and audit may be nil to disable events/audit.
func New(q Store, pub events.Publisher, a Auditor) *Handler {
	return &Handler{q: q, pub: pub, audit: a}
}

// Register mounts pick-list routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.POST("/:storeID/pick-lists", h.Create)
	g.GET("/:storeID/pick-lists", h.List)
	g.GET("/:storeID/pick-lists/:id", h.Board)
	g.POST("/:storeID/pick-lists/:id/start", h.Start)
	g.PATCH("/:storeID/pick-lists/:id/lines/:lineID", h.Pick)
	g.POST("/:storeID/pick-lists/:id/complete", h.Complete)
}

// LineRow is one stop on the pick route.
type LineRow struct {
	ID           uuid.UUID `json:"id"`
	Seq          int32     `json:"seq"`
	SKU          string    `json:"sku"`
	ZoneID       *string   `json:"zone_id,omitempty"`
	ZoneName     string    `json:"zone_name"`
	ZoneX        float64   `json:"zone_x"`
	ZoneY        float64   `json:"zone_y"`
	QtyRequested int32     `json:"qty_requested"`
	QtyPicked    int32     `json:"qty_picked"`
	Status       string    `json:"status"`
}

// ListRow is one pick list in the index.
type ListRow struct {
	ID        uuid.UUID `json:"id"`
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	LineCount int32     `json:"line_count"`
	DoneCount int32     `json:"done_count"`
	CreatedAt string    `json:"created_at"`
}

// Board is a pick list with its ordered route.
type Board struct {
	ID        uuid.UUID `json:"id"`
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	Lines     []LineRow `json:"lines"`
}

func toLineRows(rows []store.ListPickListLinesRow) []LineRow {
	out := make([]LineRow, 0, len(rows))
	for _, r := range rows {
		lr := LineRow{
			ID:           r.ID,
			Seq:          r.Seq,
			SKU:          r.Sku,
			ZoneName:     r.ZoneName,
			ZoneX:        r.ZoneX,
			ZoneY:        r.ZoneY,
			QtyRequested: r.QtyRequested,
			QtyPicked:    r.QtyPicked,
			Status:       r.Status,
		}
		if r.ZoneID.Valid {
			s := uuid.UUID(r.ZoneID.Bytes).String()
			lr.ZoneID = &s
		}
		out = append(out, lr)
	}
	return out
}

type createLine struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type createRequest struct {
	Reference  string       `json:"reference"`
	AssigneeID string       `json:"assignee_id"`
	Lines      []createLine `json:"lines"`
}

// Create godoc
//
//	@Summary	Build a pick list and optimise its route across the store map
//	@Tags		picking
//	@Accept		json
//	@Produce	json
//	@Param		storeID	path		string			true	"Store UUID"
//	@Param		body	body		createRequest	true	"Reference + lines"
//	@Success	201		{object}	Board
//	@Failure	400		{object}	map[string]string
//	@Router		/stores/{storeID}/pick-lists [post]
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
	if len(req.Lines) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "at least one line required")
	}
	for _, l := range req.Lines {
		if l.SKU == "" || l.Qty <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "each line needs a sku and qty > 0")
		}
	}

	assignee := pgtype.UUID{}
	if req.AssigneeID != "" {
		id, perr := uuid.Parse(req.AssigneeID)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid assignee_id")
		}
		assignee = pgtype.UUID{Bytes: id, Valid: true}
	}

	ctx := c.Request().Context()
	list, err := h.q.CreatePickList(ctx, store.CreatePickListParams{
		OrgID:      p.OrgID,
		StoreID:    storeID,
		Reference:  req.Reference,
		CreatedBy:  pgtype.UUID{Bytes: p.UserID, Valid: true},
		AssigneeID: assignee,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create pick list")
	}

	// Resolve each SKU to its best source zone; routable stops get optimised,
	// unresolved SKUs (no on-hand) are appended after the route.
	var stops []domain.PickStop
	var unresolved []createLine
	for _, l := range req.Lines {
		src, serr := h.q.ResolvePickSource(ctx, store.ResolvePickSourceParams{
			OrgID: p.OrgID, StoreID: storeID, Sku: l.SKU,
		})
		if serr != nil {
			if errors.Is(serr, pgx.ErrNoRows) {
				unresolved = append(unresolved, l)
				continue
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "resolve source")
		}
		stops = append(stops, domain.PickStop{
			ZoneID: src.ZoneID, SKU: l.SKU, QtyRequested: l.Qty, X: src.ZoneX, Y: src.ZoneY,
		})
	}

	seq := int32(0)
	for _, s := range domain.OptimizePickRoute(stops, 0, 0) {
		if _, err := h.q.AddPickLine(ctx, store.AddPickLineParams{
			ListID:       list.ID,
			OrgID:        p.OrgID,
			ZoneID:       pgtype.UUID{Bytes: s.ZoneID, Valid: true},
			Sku:          s.SKU,
			QtyRequested: int32(s.QtyRequested),
			Seq:          seq,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "add line")
		}
		seq++
	}
	for _, l := range unresolved {
		if _, err := h.q.AddPickLine(ctx, store.AddPickLineParams{
			ListID:       list.ID,
			OrgID:        p.OrgID,
			ZoneID:       pgtype.UUID{},
			Sku:          l.SKU,
			QtyRequested: int32(l.Qty),
			Seq:          seq,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "add line")
		}
		seq++
	}

	lines, err := h.q.ListPickListLines(ctx, list.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}
	return c.JSON(http.StatusCreated, Board{
		ID: list.ID, Reference: list.Reference, Status: list.Status, Lines: toLineRows(lines),
	})
}

// List godoc
//
//	@Summary	List pick lists for a store
//	@Tags		picking
//	@Produce	json
//	@Param		storeID	path	string	true	"Store UUID"
//	@Success	200		{array}	ListRow
//	@Router		/stores/{storeID}/pick-lists [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	rows, err := h.q.ListPickListsByStore(c.Request().Context(), store.ListPickListsByStoreParams{
		OrgID: p.OrgID, StoreID: storeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list pick lists")
	}
	out := make([]ListRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, ListRow{
			ID: r.ID, Reference: r.Reference, Status: r.Status,
			LineCount: r.LineCount, DoneCount: r.DoneCount,
			CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// Board godoc
//
//	@Summary	Get a pick list with its ordered route
//	@Tags		picking
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Pick list UUID"
//	@Success	200		{object}	Board
//	@Failure	404		{object}	map[string]string
//	@Router		/stores/{storeID}/pick-lists/{id} [get]
func (h *Handler) Board(c echo.Context) error {
	_, list, err := h.load(c)
	if err != nil {
		return err
	}
	lines, lerr := h.q.ListPickListLines(c.Request().Context(), list.ID)
	if lerr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}
	return c.JSON(http.StatusOK, Board{
		ID: list.ID, Reference: list.Reference, Status: list.Status, Lines: toLineRows(lines),
	})
}

// Start godoc
//
//	@Summary	Mark a pick list as in-progress
//	@Tags		picking
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Pick list UUID"
//	@Success	200		{object}	map[string]string
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/pick-lists/{id}/start [post]
func (h *Handler) Start(c echo.Context) error {
	p, list, err := h.load(c)
	if err != nil {
		return err
	}
	done, serr := h.q.StartPickList(c.Request().Context(), store.StartPickListParams{OrgID: p.OrgID, ID: list.ID})
	if serr != nil {
		if errors.Is(serr, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusConflict, "pick list not open")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "start pick list")
	}
	return c.JSON(http.StatusOK, map[string]string{"id": done.ID.String(), "status": done.Status})
}

type pickRequest struct {
	QtyPicked int `json:"qty_picked"`
}

// Pick godoc
//
//	@Summary	Confirm a pick for one line: decrement stock, flag short picks
//	@Tags		picking
//	@Accept		json
//	@Produce	json
//	@Param		storeID	path		string		true	"Store UUID"
//	@Param		id		path		string		true	"Pick list UUID"
//	@Param		lineID	path		string		true	"Line UUID"
//	@Param		body	body		pickRequest	true	"Quantity picked"
//	@Success	200		{object}	LineRow
//	@Failure	400		{object}	map[string]string
//	@Failure	404		{object}	map[string]string
//	@Router		/stores/{storeID}/pick-lists/{id}/lines/{lineID} [patch]
func (h *Handler) Pick(c echo.Context) error {
	p, list, err := h.load(c)
	if err != nil {
		return err
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	lineID, err := uuid.Parse(c.Param("lineID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid line id")
	}
	var req pickRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.QtyPicked < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "qty_picked must be non-negative")
	}

	ctx := c.Request().Context()
	lines, err := h.q.ListPickListLines(ctx, list.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}
	var target *store.ListPickListLinesRow
	for i := range lines {
		if lines[i].ID == lineID {
			target = &lines[i]
			break
		}
	}
	if target == nil {
		return echo.NewHTTPError(http.StatusNotFound, "line not found")
	}
	if req.QtyPicked > int(target.QtyRequested) {
		return echo.NewHTTPError(http.StatusBadRequest, "qty_picked exceeds requested")
	}

	status := domain.PickLineOutcome(int(target.QtyRequested), req.QtyPicked)
	updated, err := h.q.SetPickLinePicked(ctx, store.SetPickLinePickedParams{
		QtyPicked: int32(req.QtyPicked),
		Status:    string(status),
		OrgID:     p.OrgID,
		ID:        lineID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "set picked")
	}

	// Decrement on-hand at the source zone (floors at 0).
	if req.QtyPicked > 0 && target.ZoneID.Valid {
		zoneID := uuid.UUID(target.ZoneID.Bytes)
		if _, err := h.q.AdjustItemLocationQty(ctx, store.AdjustItemLocationQtyParams{
			OrgID: p.OrgID, StoreID: storeID, ZoneID: zoneID, Sku: target.Sku, Qty: int32(-req.QtyPicked),
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "decrement inventory")
		}
		h.record(ctx, p, target.Sku, map[string]any{
			"list_id": list.ID.String(), "zone_id": zoneID.String(),
			"qty_picked": req.QtyPicked, "qty_requested": target.QtyRequested,
		})
		if status == domain.PickLineShort {
			if err := h.maybeRestock(ctx, p.OrgID, storeID, zoneID, target.Sku); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "short-pick restock")
			}
		}
	}

	h.publishProgress(ctx, p.OrgID, storeID, list.ID, updated, lines)

	row := LineRow{
		ID: updated.ID, Seq: updated.Seq, SKU: updated.Sku,
		ZoneName: target.ZoneName, ZoneX: target.ZoneX, ZoneY: target.ZoneY,
		QtyRequested: updated.QtyRequested, QtyPicked: updated.QtyPicked, Status: updated.Status,
	}
	if target.ZoneID.Valid {
		s := uuid.UUID(target.ZoneID.Bytes).String()
		row.ZoneID = &s
	}
	return c.JSON(http.StatusOK, row)
}

// Complete godoc
//
//	@Summary	Close a pick list
//	@Tags		picking
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Pick list UUID"
//	@Success	200		{object}	map[string]any
//	@Failure	409		{object}	map[string]string
//	@Router		/stores/{storeID}/pick-lists/{id}/complete [post]
func (h *Handler) Complete(c echo.Context) error {
	p, list, err := h.load(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	done, derr := h.q.CompletePickList(ctx, store.CompletePickListParams{OrgID: p.OrgID, ID: list.ID})
	if derr != nil {
		if errors.Is(derr, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusConflict, "pick list already closed")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "complete pick list")
	}

	lines, err := h.q.ListPickListLines(ctx, list.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}
	picked, short := 0, 0
	for _, l := range lines {
		switch l.Status {
		case string(domain.PickLinePicked):
			picked++
		case string(domain.PickLineShort):
			short++
		}
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id": done.ID, "status": done.Status, "picked": picked, "short": short,
	})
}

// load resolves the principal and pick list from the request, mapping a missing
// list to 404.
func (h *Handler) load(c echo.Context) (*domain.Principal, store.PickList, error) {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return nil, store.PickList{}, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return nil, store.PickList{}, echo.NewHTTPError(http.StatusBadRequest, "invalid pick list id")
	}
	list, err := h.q.GetPickList(c.Request().Context(), store.GetPickListParams{OrgID: p.OrgID, ID: id})
	if err != nil {
		return nil, store.PickList{}, echo.NewHTTPError(http.StatusNotFound, "pick list not found")
	}
	return p, list, nil
}

// publishProgress emits a PickProgress event with the post-update completion
// counts. Best-effort: a publish failure does not fail the request.
func (h *Handler) publishProgress(
	ctx context.Context,
	orgID, storeID, listID uuid.UUID,
	updated store.PickListLine,
	lines []store.ListPickListLinesRow,
) {
	if h.pub == nil {
		return
	}
	done, total := 0, len(lines)
	for _, l := range lines {
		if l.ID == updated.ID {
			if updated.Status != string(domain.PickLinePending) {
				done++
			}
			continue
		}
		if l.Status != string(domain.PickLinePending) {
			done++
		}
	}
	_ = h.pub.Publish(ctx, events.PickProgressSubject(orgID), events.PickProgress{
		OrgID: orgID, StoreID: storeID, ListID: listID, LineID: updated.ID,
		SKU: updated.Sku, QtyPicked: int(updated.QtyPicked), Status: updated.Status,
		Done: done, Total: total, TS: time.Now().UTC(),
	})
}

// maybeRestock raises a high-priority restock task when a short pick drops a SKU
// to/below its reorder point and no open restock task already exists.
func (h *Handler) maybeRestock(ctx context.Context, orgID, storeID, zoneID uuid.UUID, sku string) error {
	if _, err := h.q.GetItemBySKU(ctx, store.GetItemBySKUParams{OrgID: orgID, Sku: sku}); err != nil {
		return nil // unknown SKU has no reorder policy
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

// record writes a pick audit entry best-effort.
func (h *Handler) record(ctx context.Context, p *domain.Principal, sku string, meta map[string]any) {
	if h.audit == nil {
		return
	}
	_ = h.audit.Write(ctx, audit.Entry{
		OrgID:        p.OrgID,
		ActorUserID:  p.UserID,
		Action:       "inventory.pick",
		ResourceType: "inventory",
		ResourceID:   sku,
		Metadata:     meta,
	})
}
