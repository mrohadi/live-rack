// Package tasks provides kanban task read + status-move endpoints.
package tasks

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/store"
)

// deadlineWindow is how close a due date must be to trigger a deadline notification.
const deadlineWindow = 24 * time.Hour

// Store is the narrow store dependency the handler needs.
type Store interface {
	CreateTask(ctx context.Context, arg store.CreateTaskParams) (store.Task, error)
	ListTasksByStore(ctx context.Context, arg store.ListTasksByStoreParams) ([]store.Task, error)
	UpdateTaskStatus(ctx context.Context, arg store.UpdateTaskStatusParams) (store.Task, error)
	AssignTask(ctx context.Context, arg store.AssignTaskParams) (store.Task, error)
}

// Handler serves task board endpoints.
type Handler struct {
	q   Store
	pub events.Publisher
}

// New creates a Handler.
func New(q Store, pub events.Publisher) *Handler {
	return &Handler{q: q, pub: pub}
}

// Register mounts task routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.POST("/:storeID/tasks", h.Create)
	g.GET("/:storeID/tasks", h.List)
	g.PATCH("/:storeID/tasks/:id", h.UpdateStatus)
	g.PATCH("/:storeID/tasks/:id/assignee", h.Assign)
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

type createRequest struct {
	ZoneID   string `json:"zone_id"`
	Title    string `json:"title"`
	Priority string `json:"priority"`
	DueAt    string `json:"due_at"`
}

// Create godoc
//
//	@Summary		Create a new task, optionally scoped to a zone
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Param			body	body		createRequest	true	"Task payload"
//	@Success		201		{object}	Row
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Router			/stores/{storeID}/tasks [post]
func (h *Handler) Create(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.CanMutateTask(p) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	storeID, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid store id")
	}

	var req createRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title required")
	}

	priority := domain.TaskPriority(req.Priority)
	switch priority {
	case domain.TaskPriorityLow, domain.TaskPriorityMed, domain.TaskPriorityHigh:
	default:
		priority = domain.TaskPriorityMed
	}

	arg := store.CreateTaskParams{
		OrgID:    p.OrgID,
		StoreID:  storeID,
		Title:    req.Title,
		Status:   string(domain.TaskStatusTodo),
		Priority: string(priority),
	}
	if req.ZoneID != "" {
		zid, err := uuid.Parse(req.ZoneID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid zone_id")
		}
		arg.ZoneID = pgtype.UUID{Bytes: zid, Valid: true}
	}
	if req.DueAt != "" {
		t, err := time.Parse(time.RFC3339, req.DueAt)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid due_at: must be RFC3339")
		}
		arg.DueAt = pgtype.Timestamptz{Time: t.UTC(), Valid: true}
	}

	task, err := h.q.CreateTask(c.Request().Context(), arg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create task")
	}
	return c.JSON(http.StatusCreated, toRow(task))
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

type assignRequest struct {
	AssigneeID uuid.UUID `json:"assignee_id"`
}

// Assign godoc
//
//	@Summary		Assign a task and notify the assignee over NATS
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Param			id		path		string			true	"Task UUID"
//	@Param			body	body		assignRequest	true	"Assignee"
//	@Success		200		{object}	Row
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/stores/{storeID}/tasks/{id}/assignee [patch]
func (h *Handler) Assign(c echo.Context) error {
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

	var req assignRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.AssigneeID == uuid.Nil {
		return echo.NewHTTPError(http.StatusBadRequest, "assignee_id required")
	}

	ctx := c.Request().Context()
	t, err := h.q.AssignTask(ctx, store.AssignTaskParams{
		OrgID:      p.OrgID,
		ID:         taskID,
		AssigneeID: pgtype.UUID{Bytes: req.AssigneeID, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	if err := h.notify(ctx, t); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "publish notification")
	}
	return c.JSON(http.StatusOK, toRow(t))
}

// notify publishes an "assigned" event, plus a "deadline" event when the task is due soon.
func (h *Handler) notify(ctx context.Context, t store.Task) error {
	var due *time.Time
	if t.DueAt.Valid {
		v := t.DueAt.Time.UTC()
		due = &v
	}
	base := events.TaskNotified{
		OrgID:      t.OrgID,
		StoreID:    t.StoreID,
		TaskID:     t.ID,
		AssigneeID: uuid.UUID(t.AssigneeID.Bytes),
		Title:      t.Title,
		DueAt:      due,
		TS:         time.Now().UTC(),
	}

	assigned := base
	assigned.Kind = events.TaskNotifyAssigned
	if err := h.pub.Publish(ctx, events.TaskSubject(t.OrgID), assigned); err != nil {
		return err
	}

	if (domain.Task{DueAt: due}).DueSoon(time.Now().UTC(), deadlineWindow) {
		deadline := base
		deadline.Kind = events.TaskNotifyDeadline
		if err := h.pub.Publish(ctx, events.TaskSubject(t.OrgID), deadline); err != nil {
			return err
		}
	}
	return nil
}
