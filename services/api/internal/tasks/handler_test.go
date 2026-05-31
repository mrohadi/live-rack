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

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/tasks"
)

type fakeStore struct {
	gotOrg    uuid.UUID
	gotStore  uuid.UUID
	gotStatus string
	rows      []store.Task
	updated   store.Task
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
	h := tasks.New(fs)
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
	h := tasks.New(fs)

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
	h := tasks.New(fs)

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

func TestTasksHandler_UpdateStatus_ReadonlyForbidden(t *testing.T) {
	orgID, storeID, taskID := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	h := tasks.New(fs)

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
