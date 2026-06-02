// Package counts serves cycle-count sessions: blind physical counts that
// reconcile variances back into on-hand inventory.
package counts

import (
	"context"
	"errors"
	"net/http"

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
	CreateCycleCount(ctx context.Context, arg store.CreateCycleCountParams) (store.CycleCount, error)
	SnapshotCountLines(ctx context.Context, arg store.SnapshotCountLinesParams) error
	GetCycleCount(ctx context.Context, arg store.GetCycleCountParams) (store.CycleCount, error)
	ListCountLines(ctx context.Context, countID uuid.UUID) ([]store.CycleCountLine, error)
	SetCountedQty(ctx context.Context, arg store.SetCountedQtyParams) (store.CycleCountLine, error)
	CompleteCycleCount(ctx context.Context, arg store.CompleteCycleCountParams) (store.CycleCount, error)
	SetItemLocationQty(ctx context.Context, arg store.SetItemLocationQtyParams) (store.ItemLocation, error)
}

// Auditor records an append-only audit entry. *audit.Writer satisfies it.
type Auditor interface {
	Write(ctx context.Context, e audit.Entry) error
}

// Handler serves cycle-count endpoints.
type Handler struct {
	q     Store
	audit Auditor
}

// New creates a Handler. audit may be nil to disable audit logging.
func New(q Store, a Auditor) *Handler {
	return &Handler{q: q, audit: a}
}

// Register mounts count routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.POST("/:storeID/counts", h.Create)
	g.GET("/:storeID/counts/:id", h.Get)
	g.PATCH("/:storeID/counts/:id/lines", h.SetLine)
	g.POST("/:storeID/counts/:id/complete", h.Complete)
}

// Session is the count session returned to the client.
type Session struct {
	ID     uuid.UUID `json:"id"`
	ZoneID uuid.UUID `json:"zone_id"`
	Status string    `json:"status"`
	Lines  []Line    `json:"lines"`
}

// Line is one SKU in a count. SystemQty is hidden (0) while the count is open
// so the physical count stays blind; it is revealed on completion.
type Line struct {
	SKU        string `json:"sku"`
	SystemQty  int32  `json:"system_qty"`
	CountedQty *int32 `json:"counted_qty,omitempty"`
}

func toLines(rows []store.CycleCountLine, blind bool) []Line {
	out := make([]Line, 0, len(rows))
	for _, r := range rows {
		l := Line{SKU: r.Sku}
		if !blind {
			l.SystemQty = r.SystemQty
		}
		if r.CountedQty.Valid {
			v := r.CountedQty.Int32
			l.CountedQty = &v
		}
		out = append(out, l)
	}
	return out
}

type createRequest struct {
	ZoneID string `json:"zone_id"`
}

// Create godoc
//
//	@Summary		Start a cycle-count session for a zone (snapshots on-hand)
//	@Tags			counts
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Param			body	body		createRequest	true	"Zone"
//	@Success		201		{object}	Session
//	@Failure		400		{object}	map[string]string
//	@Router			/stores/{storeID}/counts [post]
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
	zoneID, err := uuid.Parse(req.ZoneID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid zone_id")
	}

	ctx := c.Request().Context()
	cc, err := h.q.CreateCycleCount(ctx, store.CreateCycleCountParams{
		OrgID:     p.OrgID,
		StoreID:   storeID,
		ZoneID:    zoneID,
		CreatedBy: pgtype.UUID{Bytes: p.UserID, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create count")
	}

	if err := h.q.SnapshotCountLines(ctx, store.SnapshotCountLinesParams{
		CountID: cc.ID,
		OrgID:   p.OrgID,
		StoreID: storeID,
		ZoneID:  zoneID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "snapshot lines")
	}

	lines, err := h.q.ListCountLines(ctx, cc.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}

	return c.JSON(http.StatusCreated, Session{
		ID:     cc.ID,
		ZoneID: cc.ZoneID,
		Status: cc.Status,
		Lines:  toLines(lines, true),
	})
}

// Get godoc
//
//	@Summary		Fetch a cycle-count session and its lines
//	@Tags			counts
//	@Produce		json
//	@Param			storeID	path		string	true	"Store UUID"
//	@Param			id		path		string	true	"Count UUID"
//	@Success		200		{object}	Session
//	@Failure		404		{object}	map[string]string
//	@Router			/stores/{storeID}/counts/{id} [get]
func (h *Handler) Get(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	countID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid count id")
	}

	ctx := c.Request().Context()
	cc, err := h.q.GetCycleCount(ctx, store.GetCycleCountParams{OrgID: p.OrgID, ID: countID})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "count not found")
	}
	lines, err := h.q.ListCountLines(ctx, cc.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}

	blind := cc.Status == string(domain.CycleCountOpen)
	return c.JSON(http.StatusOK, Session{
		ID:     cc.ID,
		ZoneID: cc.ZoneID,
		Status: cc.Status,
		Lines:  toLines(lines, blind),
	})
}

type setLineRequest struct {
	SKU        string `json:"sku"`
	CountedQty int32  `json:"counted_qty"`
}

// SetLine godoc
//
//	@Summary		Record the blind physical count for one SKU
//	@Tags			counts
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Param			id		path		string			true	"Count UUID"
//	@Param			body	body		setLineRequest	true	"SKU + counted qty"
//	@Success		200		{object}	map[string]any
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/stores/{storeID}/counts/{id}/lines [patch]
func (h *Handler) SetLine(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	countID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid count id")
	}

	var req setLineRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.SKU == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sku required")
	}
	if req.CountedQty < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "counted_qty must be non-negative")
	}

	line, err := h.q.SetCountedQty(c.Request().Context(), store.SetCountedQtyParams{
		OrgID:      p.OrgID,
		CountID:    countID,
		Sku:        req.SKU,
		CountedQty: req.CountedQty,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "line not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "set counted qty")
	}

	return c.JSON(http.StatusOK, map[string]any{"sku": line.Sku, "counted_qty": req.CountedQty})
}

// VarianceRow is one reconciled line in the completion summary.
type VarianceRow struct {
	SKU        string `json:"sku"`
	SystemQty  int32  `json:"system_qty"`
	CountedQty int32  `json:"counted_qty"`
	Variance   int    `json:"variance"`
}

// CompleteResponse summarises the reconciliation.
type CompleteResponse struct {
	ID         uuid.UUID     `json:"id"`
	Status     string        `json:"status"`
	Reconciled int           `json:"reconciled"`
	Variances  []VarianceRow `json:"variances"`
}

// Complete godoc
//
//	@Summary		Reconcile a count: apply counted qty, record variances
//	@Tags			counts
//	@Produce		json
//	@Param			storeID	path		string	true	"Store UUID"
//	@Param			id		path		string	true	"Count UUID"
//	@Success		200		{object}	CompleteResponse
//	@Failure		404		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Router			/stores/{storeID}/counts/{id}/complete [post]
func (h *Handler) Complete(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}
	countID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid count id")
	}

	ctx := c.Request().Context()
	cc, err := h.q.GetCycleCount(ctx, store.GetCycleCountParams{OrgID: p.OrgID, ID: countID})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "count not found")
	}
	if cc.Status != string(domain.CycleCountOpen) {
		return echo.NewHTTPError(http.StatusConflict, "count already completed")
	}

	lines, err := h.q.ListCountLines(ctx, cc.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list lines")
	}

	variances := make([]VarianceRow, 0)
	for _, l := range lines {
		if !l.CountedQty.Valid {
			continue // unentered line — leave system qty untouched
		}
		counted := l.CountedQty.Int32
		v := domain.CountVariance(int(l.SystemQty), int(counted))
		if v == 0 {
			continue // no correction needed
		}

		if _, err := h.q.SetItemLocationQty(ctx, store.SetItemLocationQtyParams{
			OrgID:   p.OrgID,
			StoreID: storeID,
			ZoneID:  cc.ZoneID,
			Sku:     l.Sku,
			Qty:     counted,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "apply count")
		}

		h.record(ctx, p, l.Sku, map[string]any{
			"count_id": cc.ID.String(), "zone_id": cc.ZoneID.String(),
			"system_qty": l.SystemQty, "counted_qty": counted, "variance": v,
		})

		variances = append(variances, VarianceRow{
			SKU:        l.Sku,
			SystemQty:  l.SystemQty,
			CountedQty: counted,
			Variance:   v,
		})
	}

	done, err := h.q.CompleteCycleCount(ctx, store.CompleteCycleCountParams{OrgID: p.OrgID, ID: countID})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "complete count")
	}

	return c.JSON(http.StatusOK, CompleteResponse{
		ID:         done.ID,
		Status:     done.Status,
		Reconciled: len(variances),
		Variances:  variances,
	})
}

// record writes a cycle-count audit entry best-effort.
func (h *Handler) record(ctx context.Context, p *domain.Principal, sku string, meta map[string]any) {
	if h.audit == nil {
		return
	}
	_ = h.audit.Write(ctx, audit.Entry{
		OrgID:        p.OrgID,
		ActorUserID:  p.UserID,
		Action:       "inventory.cycle_count",
		ResourceType: "inventory",
		ResourceID:   sku,
		Metadata:     meta,
	})
}
