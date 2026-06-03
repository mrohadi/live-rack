package stores_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	storesapi "github.com/live-rack/services/api/internal/stores"
)

var (
	orgID  = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	storeA = uuid.MustParse("00000000-0000-0000-0000-000000000010")
)

type fakeQ struct {
	rows      []store.Store
	createErr error
	created   store.CreateStoreParams
}

func (f *fakeQ) ListStoresByOrg(_ context.Context, _ uuid.UUID) ([]store.Store, error) {
	return f.rows, nil
}
func (f *fakeQ) CreateStore(_ context.Context, arg store.CreateStoreParams) (store.Store, error) {
	if f.createErr != nil {
		return store.Store{}, f.createErr
	}
	f.created = arg
	return store.Store{ID: storeA, OrgID: orgID, Name: arg.Name, Timezone: arg.Timezone}, nil
}

func makePrincipal(role domain.RoleName) *domain.Principal {
	return &domain.Principal{UserID: userID, OrgID: orgID, IDPOrgID: "idp-org", Role: role}
}

func serve(t *testing.T, q *fakeQ, method, path, body string, p *domain.Principal) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	g := e.Group("/api/v1", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := pkgauth.WithPrincipal(c.Request().Context(), p)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	storesapi.New(q).Register(g)
	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestList_ReturnsOrgStores(t *testing.T) {
	q := &fakeQ{rows: []store.Store{
		{ID: storeA, OrgID: orgID, Name: "Main Warehouse", Timezone: "UTC"},
	}}
	rec := serve(t, q, http.MethodGet, "/api/v1/stores", "", makePrincipal(domain.RoleAdmin))
	require.Equal(t, http.StatusOK, rec.Code)
	var out []storesapi.StoreResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "Main Warehouse", out[0].Name)
}

func TestCreate_AdminCreatesStore(t *testing.T) {
	q := &fakeQ{}
	rec := serve(t, q, http.MethodPost, "/api/v1/stores",
		`{"name":"Depot B","timezone":"Asia/Jakarta"}`, makePrincipal(domain.RoleAdmin))
	require.Equal(t, http.StatusCreated, rec.Code)
	var out storesapi.StoreResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "Depot B", out.Name)
	assert.Equal(t, "Depot B", q.created.Name)
}

func TestCreate_NonAdminForbidden(t *testing.T) {
	q := &fakeQ{}
	rec := serve(t, q, http.MethodPost, "/api/v1/stores",
		`{"name":"X"}`, makePrincipal(domain.RoleStaff))
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreate_MissingNameBadRequest(t *testing.T) {
	q := &fakeQ{}
	rec := serve(t, q, http.MethodPost, "/api/v1/stores", `{}`, makePrincipal(domain.RoleAdmin))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// pgtype.Text is referenced in handler — ensure package compiles.
var _ = pgtype.Text{}
