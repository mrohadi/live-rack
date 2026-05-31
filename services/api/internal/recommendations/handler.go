// Package recommendations turns an Analytics-screen recommendation into a real
// task ("Apply"). The recommendation itself is delivered to the client over the
// WebSocket signal stream; this endpoint persists the operator's decision.
package recommendations

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/store"
)

// defaults for tasks created from a recommendation.
const (
	taskStatus   = "todo"
	taskPriority = "high"
)

// TaskCreator persists a new task. *store.Queries satisfies it.
type TaskCreator interface {
	CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error)
}

// Handler serves the Apply endpoint.
type Handler struct {
	q TaskCreator
}

// New creates a Handler.
func New(q TaskCreator) *Handler {
	return &Handler{q: q}
}

// Register mounts recommendation routes on the authenticated API group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/recommendations/apply", h.Apply)
}

// ApplyRequest applies a recommendation by creating a task for its store.
type ApplyRequest struct {
	StoreID       string `json:"store_id"`
	SuggestedTask string `json:"suggested_task"`
}

// ApplyResponse returns the created task id.
type ApplyResponse struct {
	TaskID uuid.UUID `json:"task_id"`
	Status string    `json:"status"`
}

// Apply godoc
//
//	@Summary	Create a task from a recommendation's suggested action
//	@Tags		recommendations
//	@Accept		json
//	@Produce	json
//	@Param		body	body		ApplyRequest	true	"Recommendation to apply"
//	@Success	201		{object}	ApplyResponse
//	@Router		/recommendations/apply [post]
func (h *Handler) Apply(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req ApplyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	storeID, err := uuid.Parse(req.StoreID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store_id")
	}
	title := strings.TrimSpace(req.SuggestedTask)
	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "suggested_task required")
	}

	task, err := h.q.CreateTask(c.Request().Context(), store.CreateTaskParams{
		OrgID:    p.OrgID,
		StoreID:  storeID,
		ZoneID:   pgtype.UUID{Valid: false},
		Title:    title,
		Status:   taskStatus,
		Priority: taskPriority,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create task")
	}
	return c.JSON(http.StatusCreated, ApplyResponse{TaskID: task.ID, Status: task.Status})
}
