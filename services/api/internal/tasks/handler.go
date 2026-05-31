// Package tasks provides kanban task read + status-move endpoints.
package tasks

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
	ListTasksByStore(ctx context.Context, arg store.ListTasksByStoreParams) ([]store.Task, error)
	UpdateTaskStatus(ctx context.Context, arg store.UpdateTaskStatusParams) (store.Task, error)
}

// Handler serves task board endpoints.
type Handler struct {
	q Store
}

// New creates a Handler.
func New(q Store) *Handler {
	return &Handler{q: q}
}

// Register mounts task routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.GET("/:storeID/tasks", h.List)
	g.PATCH("/:storeID/tasks/:id", h.UpdateStatus)
}

// Row is one kanban card returned to the client.
type Row struct {
	ID         uuid.UUID `json:"id"`
	StoreID    uuid.UUID `json:"store_id"`
	ZoneID     *string   `json:"zone_id,omitempty"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	Priority   string    `json:"priority"`
	AssigneeID *string   `json:"assignee_id,omitempty"`
	DueAt      *string   `json:"due_at,omitempty"`
	UpdatedAt  string    `json:"updated_at"`
}

func toRow(t store.Task) Row {
	r := Row{
		ID:        t.ID,
		StoreID:   t.StoreID,
		Title:     t.Title,
		Status:    t.Status,
		Priority:  t.Priority,
		UpdatedAt: t.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if t.ZoneID.Valid {
		s := uuid.UUID(t.ZoneID.Bytes).String()
		r.ZoneID = &s
	}
	if t.AssigneeID.Valid {
		s := uuid.UUID(t.AssigneeID.Bytes).String()
		r.AssigneeID = &s
	}
	if t.DueAt.Valid {
		s := t.DueAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		r.DueAt = &s
	}
	return r
}

// List godoc
//
//	@Summary		List tasks for a store, grouped client-side into kanban columns
//	@Tags			tasks
//	@Produce		json
//	@Param			storeID	path		string	true	"Store UUID"
//	@Success		200		{array}		Row
//	@Failure		400		{object}	map[string]string
//	@Router			/stores/{storeID}/tasks [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}

	rows, err := h.q.ListTasksByStore(c.Request().Context(), store.ListTasksByStoreParams{
		OrgID:   p.OrgID,
		StoreID: storeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list tasks")
	}

	out := make([]Row, 0, len(rows))
	for _, t := range rows {
		out = append(out, toRow(t))
	}
	return c.JSON(http.StatusOK, out)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateStatus godoc
//
//	@Summary		Move a task to a new kanban column
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string				true	"Store UUID"
//	@Param			id		path		string				true	"Task UUID"
//	@Param			body	body		updateStatusRequest	true	"New status"
//	@Success		200		{object}	Row
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/stores/{storeID}/tasks/{id} [patch]
func (h *Handler) UpdateStatus(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.CanMutateTask(p) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	var req updateStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if !domain.TaskStatus(req.Status).Valid() {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}

	t, err := h.q.UpdateTaskStatus(c.Request().Context(), store.UpdateTaskStatusParams{
		OrgID:  p.OrgID,
		ID:     taskID,
		Status: req.Status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}
	return c.JSON(http.StatusOK, toRow(t))
}
