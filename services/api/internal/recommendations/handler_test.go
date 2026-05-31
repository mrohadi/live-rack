package recommendations_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/recommendations"
)

type fakeCreator struct {
	got store.CreateTaskParams
	err error
}

func (f *fakeCreator) CreateTask(_ context.Context, arg store.CreateTaskParams) (store.Task, error) {
	f.got = arg
	if f.err != nil {
		return store.Task{}, f.err
	}
	return store.Task{ID: uuid.New(), Status: arg.Status}, nil
}

func do(t *testing.T, fc *fakeCreator, org uuid.UUID, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	recommendations.New(fc).Register(e.Group("/api/v1"))
	p := &domain.Principal{UserID: uuid.New(), OrgID: org, Role: domain.RoleManager}
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodPost, "/api/v1/recommendations/apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestApply_CreatesTask(t *testing.T) {
	org, storeID := uuid.New(), uuid.New()
	fc := &fakeCreator{}
	rec := do(t, fc, org, `{"store_id":"`+storeID.String()+`","suggested_task":"Stock umbrellas"}`)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, org, fc.got.OrgID)
	assert.Equal(t, storeID, fc.got.StoreID)
	assert.Equal(t, "Stock umbrellas", fc.got.Title)
	assert.Equal(t, "todo", fc.got.Status)
	assert.Equal(t, "high", fc.got.Priority)

	var out recommendations.ApplyResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "todo", out.Status)
}

func TestApply_BadStore(t *testing.T) {
	rec := do(t, &fakeCreator{}, uuid.New(), `{"store_id":"nope","suggested_task":"x"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestApply_EmptyTask(t *testing.T) {
	rec := do(t, &fakeCreator{}, uuid.New(), `{"store_id":"`+uuid.New().String()+`","suggested_task":"  "}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
