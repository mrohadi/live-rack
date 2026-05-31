// Package pipelines serves pipeline board read + card-move endpoints.
package pipelines

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// Store is the narrow store dependency the handler needs.
type Store interface {
	ListPipelinesByStore(ctx context.Context, arg store.ListPipelinesByStoreParams) ([]store.Pipeline, error)
	GetPipeline(ctx context.Context, arg store.GetPipelineParams) (store.Pipeline, error)
	ListStagesByPipeline(ctx context.Context, arg store.ListStagesByPipelineParams) ([]store.PipelineStage, error)
	ListCardsByPipeline(ctx context.Context, arg store.ListCardsByPipelineParams) ([]store.PipelineCard, error)
	MoveCard(ctx context.Context, arg store.MoveCardParams) (store.PipelineCard, error)
}

// Handler serves pipeline endpoints.
type Handler struct {
	q Store
}

// New creates a Handler.
func New(q Store) *Handler {
	return &Handler{q: q}
}

// Register mounts pipeline routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.GET("/:storeID/pipelines", h.List)
	g.GET("/:storeID/pipelines/:id", h.Board)
	g.PATCH("/:storeID/pipelines/:id/cards/:cardID", h.MoveCard)
}

// PipelineRow is one pipeline in the selector list.
type PipelineRow struct {
	ID   uuid.UUID `json:"id"`
	Key  string    `json:"key"`
	Name string    `json:"name"`
}

// StageRow is one column of the board.
type StageRow struct {
	Position   int32  `json:"position"`
	Name       string `json:"name"`
	SLASeconds int64  `json:"sla_seconds"`
}

// CardRow is one card on the board. AgeSeconds + Ageing are derived server-side
// so every client renders ageing alerts identically.
type CardRow struct {
	ID             uuid.UUID `json:"id"`
	StagePosition  int32     `json:"stage_position"`
	Title          string    `json:"title"`
	SKU            string    `json:"sku,omitempty"`
	Priority       string    `json:"priority"`
	OwnerID        *string   `json:"owner_id,omitempty"`
	EnteredStageAt string    `json:"entered_stage_at"`
	AgeSeconds     int64     `json:"age_seconds"`
	Ageing         bool      `json:"ageing"`
}

// BoardResponse is the full board for one pipeline.
type BoardResponse struct {
	Pipeline PipelineRow `json:"pipeline"`
	Stages   []StageRow  `json:"stages"`
	Cards    []CardRow   `json:"cards"`
}

func rfc3339(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z07:00")
}

// List godoc
//
//	@Summary	List pipelines for a store
//	@Tags		pipelines
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Success	200		{array}		PipelineRow
//	@Router		/stores/{storeID}/pipelines [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}

	rows, err := h.q.ListPipelinesByStore(c.Request().Context(), store.ListPipelinesByStoreParams{
		OrgID: p.OrgID, StoreID: storeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list pipelines")
	}
	out := make([]PipelineRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, PipelineRow{ID: r.ID, Key: r.Key, Name: r.Name})
	}
	return c.JSON(http.StatusOK, out)
}

// Board godoc
//
//	@Summary	Get a pipeline board — stages + cards with ageing flags
//	@Tags		pipelines
//	@Produce	json
//	@Param		storeID	path		string	true	"Store UUID"
//	@Param		id		path		string	true	"Pipeline UUID"
//	@Success	200		{object}	BoardResponse
//	@Failure	404		{object}	map[string]string
//	@Router		/stores/{storeID}/pipelines/{id} [get]
func (h *Handler) Board(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	pipeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pipeline id")
	}
	ctx := c.Request().Context()

	pipe, err := h.q.GetPipeline(ctx, store.GetPipelineParams{OrgID: p.OrgID, ID: pipeID})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "pipeline not found")
	}
	stageRows, err := h.q.ListStagesByPipeline(ctx, store.ListStagesByPipelineParams{OrgID: p.OrgID, PipelineID: pipeID})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list stages")
	}
	cardRows, err := h.q.ListCardsByPipeline(ctx, store.ListCardsByPipelineParams{OrgID: p.OrgID, PipelineID: pipeID})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list cards")
	}

	return c.JSON(http.StatusOK, buildBoard(pipe, stageRows, cardRows, time.Now().UTC()))
}

// buildBoard assembles the board response and computes ageing per card against
// the SLA of the stage the card currently sits in.
func buildBoard(pipe store.Pipeline, stages []store.PipelineStage, cards []store.PipelineCard, now time.Time) BoardResponse {
	slaByPos := make(map[int32]int64, len(stages))
	stageRows := make([]StageRow, 0, len(stages))
	for _, s := range stages {
		slaByPos[s.Position] = s.SlaSeconds
		stageRows = append(stageRows, StageRow{Position: s.Position, Name: s.Name, SLASeconds: s.SlaSeconds})
	}

	cardRows := make([]CardRow, 0, len(cards))
	for _, cd := range cards {
		age := int64(now.Sub(cd.EnteredStageAt).Seconds())
		if age < 0 {
			age = 0
		}
		sla := slaByPos[cd.StagePosition]
		row := CardRow{
			ID:             cd.ID,
			StagePosition:  cd.StagePosition,
			Title:          cd.Title,
			SKU:            cd.Sku,
			Priority:       cd.Priority,
			EnteredStageAt: rfc3339(cd.EnteredStageAt),
			AgeSeconds:     age,
			Ageing:         sla > 0 && age > sla,
		}
		if cd.OwnerID.Valid {
			s := uuid.UUID(cd.OwnerID.Bytes).String()
			row.OwnerID = &s
		}
		cardRows = append(cardRows, row)
	}

	return BoardResponse{
		Pipeline: PipelineRow{ID: pipe.ID, Key: pipe.Key, Name: pipe.Name},
		Stages:   stageRows,
		Cards:    cardRows,
	}
}

type moveCardRequest struct {
	StagePosition int32 `json:"stage_position"`
}

// MoveCard godoc
//
//	@Summary	Move a card to a different stage
//	@Tags		pipelines
//	@Accept		json
//	@Produce	json
//	@Param		storeID	path		string			true	"Store UUID"
//	@Param		id		path		string			true	"Pipeline UUID"
//	@Param		cardID	path		string			true	"Card UUID"
//	@Param		body	body		moveCardRequest	true	"Target stage"
//	@Success	200		{object}	CardRow
//	@Failure	400		{object}	map[string]string
//	@Failure	403		{object}	map[string]string
//	@Failure	404		{object}	map[string]string
//	@Router		/stores/{storeID}/pipelines/{id}/cards/{cardID} [patch]
func (h *Handler) MoveCard(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.CanMutatePipeline(p) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	cardID, err := uuid.Parse(c.Param("cardID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid card id")
	}
	var req moveCardRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.StagePosition < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "stage_position must be >= 0")
	}

	cd, err := h.q.MoveCard(c.Request().Context(), store.MoveCardParams{
		OrgID: p.OrgID, ID: cardID, StagePosition: req.StagePosition,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "card not found")
	}

	row := CardRow{
		ID:             cd.ID,
		StagePosition:  cd.StagePosition,
		Title:          cd.Title,
		SKU:            cd.Sku,
		Priority:       cd.Priority,
		EnteredStageAt: rfc3339(cd.EnteredStageAt),
		AgeSeconds:     0,
		Ageing:         false,
	}
	if cd.OwnerID.Valid {
		s := uuid.UUID(cd.OwnerID.Bytes).String()
		row.OwnerID = &s
	}
	return c.JSON(http.StatusOK, row)
}
