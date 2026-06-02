package tasks_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v5/pgtype"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/tasks"
)

type fakeStore struct {
	gotOrg      uuid.UUID
	gotStore    uuid.UUID
	gotStatus   string
	gotAssignee uuid.UUID
	rows        []store.Task
	updated     store.Task
	assigned    store.Task
	created     store.Task
	createArg   store.CreateTaskParams
}

func (f *fakeStore) CreateTask(_ context.Context, arg store.CreateTaskParams) (store.Task, error) {
	f.createArg = arg
	f.created.ID = uuid.New()
	f.created.OrgID = arg.OrgID
	f.created.StoreID = arg.StoreID
	f.created.ZoneID = arg.ZoneID
	f.created.Title = arg.Title
	f.created.Status = arg.Status
	f.created.Priority = arg.Priority
	f.created.AssigneeID = arg.AssigneeID
	f.created.DueAt = arg.DueAt
	f.created.UpdatedAt = time.Now().UTC()
	return f.created, nil
}

func (f *fakeStore) ListTasksByStore(_ context.Context, arg store.ListTasksByStoreParams) ([]store.Task, error) {
	f.gotOrg = arg.OrgID
	f.gotStore = arg.StoreID
	return f.rows, nil
}

func (f *fakeStore) UpdateTaskStatus(_ context.Context, arg store.UpdateTaskStatusParams) (store.Task, error) {
	f.gotOrg = arg.OrgID
	f.gotStatus = arg.Status
	f.updated.Status = arg.Status
	return f.updated, nil
}

func (f *fakeStore) AssignTask(_ context.Context, arg store.AssignTaskParams) (store.Task, error) {
	f.gotOrg = arg.OrgID
	f.gotAssignee = uuid.UUID(arg.AssigneeID.Bytes)
	f.assigned.AssigneeID = arg.AssigneeID
	return f.assigned, nil
}

type fakePublisher struct {
	subjects []string
	payloads []any
}

func (p *fakePublisher) Publish(_ context.Context, subject string, v any) error {
	p.subjects = append(p.subjects, subject)
	p.payloads = append(p.payloads, v)
	return nil
}

func newContext(t *testing.T, e *echo.Echo, method, target, body string, p *domain.Principal) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	var rdr *strings.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	} else {
		rdr = strings.NewReader("")
	}
	req := httptest.NewRequestWithContext(context.Background(), method, target, rdr)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetRequest(c.Request().WithContext(pkgauth.WithPrincipal(c.Request().Context(), p)))
	return c, rec
}

func TestTasksHandler_List(t *testing.T) {
	orgID, storeID, zoneID := uuid.New(), uuid.New(), uuid.New()
	row := store.Task{
		ID: uuid.New(), OrgID: orgID, StoreID: storeID,
		Title: "Restock A1", Status: "todo", Priority: "high",
		UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, row.ZoneID.Scan(zoneID.String()))

	fs := &fakeStore{rows: []store.Task{row}}
	e := echo.New()
	e.HideBanner = true
	h := tasks.New(fs, &fakePublisher{})
	h.Register(e.Group("/api/v1/stores"))

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodGet, "/api/v1/stores/"+storeID.String()+"/tasks", "", p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, orgID, fs.gotOrg)
	assert.Equal(t, storeID, fs.gotStore)

	var out []tasks.Row
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "Restock A1", out[0].Title)
	assert.Equal(t, "todo", out[0].Status)
	require.NotNil(t, out[0].ZoneID)
	assert.Equal(t, zoneID.String(), *out[0].ZoneID)
}

func TestTasksHandler_UpdateStatus(t *testing.T) {
	orgID, storeID, taskID := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{updated: store.Task{ID: taskID, OrgID: orgID, StoreID: storeID, Title: "Move me", Priority: "med", UpdatedAt: time.Now().UTC()}}
	e := echo.New()
	e.HideBanner = true
	h := tasks.New(fs, &fakePublisher{})

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/tasks/"+taskID.String(), `{"status":"done"}`, p)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), taskID.String())

	require.NoError(t, h.UpdateStatus(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "done", fs.gotStatus)

	var out tasks.Row
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "done", out.Status)
}

func TestTasksHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	orgID, storeID, taskID := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	h := tasks.New(fs, &fakePublisher{})

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, _ := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/tasks/"+taskID.String(), `{"status":"bogus"}`, p)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), taskID.String())

	err := h.UpdateStatus(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Empty(t, fs.gotStatus)
}

func TestTasksHandler_Assign_NotifiesAssignee(t *testing.T) {
	orgID, storeID, taskID, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	due := time.Now().UTC().Add(3 * time.Hour) // within deadline window
	fs := &fakeStore{assigned: store.Task{
		ID: taskID, OrgID: orgID, StoreID: storeID, Title: "Restock A1",
		Priority: "high", DueAt: pgtype.Timestamptz{Time: due, Valid: true},
		UpdatedAt: time.Now().UTC(),
	}}
	pub := &fakePublisher{}
	e := echo.New()
	h := tasks.New(fs, pub)

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleManager}
	c, rec := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/tasks/"+taskID.String()+"/assignee",
		`{"assignee_id":"`+userID.String()+`"}`, p)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), taskID.String())

	require.NoError(t, h.Assign(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, userID, fs.gotAssignee)

	// Due soon → both assigned and deadline notifications fire.
	require.Len(t, pub.payloads, 2)
	assert.Equal(t, events.TaskSubject(orgID), pub.subjects[0])
	n0 := pub.payloads[0].(events.TaskNotified)
	assert.Equal(t, events.TaskNotifyAssigned, n0.Kind)
	assert.Equal(t, userID, n0.AssigneeID)
	n1 := pub.payloads[1].(events.TaskNotified)
	assert.Equal(t, events.TaskNotifyDeadline, n1.Kind)
}

func TestTasksHandler_Assign_NoDeadlineWhenFarOff(t *testing.T) {
	orgID, storeID, taskID, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	due := time.Now().UTC().Add(72 * time.Hour) // outside window
	fs := &fakeStore{assigned: store.Task{
		ID: taskID, OrgID: orgID, StoreID: storeID, Title: "Later",
		Priority: "low", DueAt: pgtype.Timestamptz{Time: due, Valid: true},
		UpdatedAt: time.Now().UTC(),
	}}
	pub := &fakePublisher{}
	e := echo.New()
	h := tasks.New(fs, pub)

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleManager}
	c, _ := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/tasks/"+taskID.String()+"/assignee",
		`{"assignee_id":"`+userID.String()+`"}`, p)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), taskID.String())

	require.NoError(t, h.Assign(c))
	require.Len(t, pub.payloads, 1)
	assert.Equal(t, events.TaskNotifyAssigned, pub.payloads[0].(events.TaskNotified).Kind)
}

func TestTasksHandler_Assign_ReadonlyForbidden(t *testing.T) {
	orgID, storeID, taskID, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{}
	pub := &fakePublisher{}
	e := echo.New()
	h := tasks.New(fs, pub)

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleReadonly}
	c, _ := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/tasks/"+taskID.String()+"/assignee",
		`{"assignee_id":"`+userID.String()+`"}`, p)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), taskID.String())

	err := h.Assign(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, he.Code)
	assert.Empty(t, pub.payloads)
}

func TestTasksHandler_UpdateStatus_ReadonlyForbidden(t *testing.T) {
	orgID, storeID, taskID := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	h := tasks.New(fs, &fakePublisher{})

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleReadonly}
	c, _ := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/tasks/"+taskID.String(), `{"status":"done"}`, p)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), taskID.String())

	err := h.UpdateStatus(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, he.Code)
}

func TestTasksHandler_Create(t *testing.T) {
	orgID, storeID, zoneID := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	e.HideBanner = true
	h := tasks.New(fs, &fakePublisher{})

	body := `{"zone_id":"` + zoneID.String() + `","title":"Restock frozen","priority":"high","due_at":"2026-12-01T09:00:00Z"}`
	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/tasks", body, p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	assert.Equal(t, orgID, fs.createArg.OrgID)
	assert.Equal(t, storeID, fs.createArg.StoreID)
	assert.Equal(t, "Restock frozen", fs.createArg.Title)
	assert.Equal(t, "todo", fs.createArg.Status)
	assert.Equal(t, "high", fs.createArg.Priority)
	assert.True(t, fs.createArg.ZoneID.Valid)
	assert.Equal(t, zoneID, uuid.UUID(fs.createArg.ZoneID.Bytes))
	assert.True(t, fs.createArg.DueAt.Valid)

	var out tasks.Row
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "Restock frozen", out.Title)
	assert.Equal(t, "todo", out.Status)
	assert.Equal(t, "high", out.Priority)
}

func TestTasksHandler_Create_NotifiesAssignee(t *testing.T) {
	orgID, storeID, assignee := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{}
	pub := &fakePublisher{}
	e := echo.New()
	e.HideBanner = true
	h := tasks.New(fs, pub)

	body := `{"title":"Restock","priority":"high","assignee_id":"` + assignee.String() + `"}`
	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/tasks", body, p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.True(t, fs.createArg.AssigneeID.Valid)
	require.Len(t, pub.subjects, 1)
	assert.Equal(t, events.TaskSubject(orgID), pub.subjects[0])
}

func TestTasksHandler_Create_NoAssignee_NoNotify(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{}
	pub := &fakePublisher{}
	e := echo.New()
	h := tasks.New(fs, pub)

	body := `{"title":"Restock","priority":"high"}`
	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/tasks", body, p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Empty(t, pub.subjects)
}

func TestTasksHandler_Create_DefaultPriority(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	h := tasks.New(fs, &fakePublisher{})

	body := `{"title":"Quick check","priority":"bogus"}`
	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/tasks", body, p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
	// Invalid priority falls back to med.
	assert.Equal(t, "med", fs.createArg.Priority)
}

func TestTasksHandler_Create_MissingTitle(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	h := tasks.New(fs, &fakePublisher{})

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, _ := newContext(t, e, http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/tasks", `{"priority":"high"}`, p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestTasksHandler_Create_ReadonlyForbidden(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	h := tasks.New(fs, &fakePublisher{})

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleReadonly}
	c, _ := newContext(t, e, http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/tasks", `{"title":"X","priority":"low"}`, p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, he.Code)
}
