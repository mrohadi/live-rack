package users_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/users"
)

type fakeStore struct {
	rows   []store.UserListRow
	gotOrg uuid.UUID
}

func (f *fakeStore) ListUsersByOrg(_ context.Context, orgID uuid.UUID) ([]store.UserListRow, error) {
	f.gotOrg = orgID
	return f.rows, nil
}

func serve(t *testing.T, p *domain.Principal, target string) (*httptest.ResponseRecorder, *fakeStore) {
	t.Helper()
	fs := &fakeStore{rows: []store.UserListRow{{ID: uuid.New(), Email: "a@b.io", DisplayName: "Ann", Role: "admin"}}}
	e := echo.New()
	users.New(fs).Register(e.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec, fs
}

func TestMe_Capabilities(t *testing.T) {
	org := uuid.New()
	p := &domain.Principal{UserID: uuid.New(), OrgID: org, Role: domain.RoleManager, MFAVerified: true, ZoneIDs: []uuid.UUID{uuid.New()}}
	rec, _ := serve(t, p, "/api/v1/me")
	require.Equal(t, http.StatusOK, rec.Code)

	var out users.CapabilitiesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "manager", out.Role)
	assert.True(t, out.MFAVerified)
	assert.True(t, out.ZoneScoped)
	assert.False(t, out.StoreScoped)
	assert.Contains(t, out.Permissions, "edit_zones")
	assert.NotContains(t, out.Permissions, "edit_users")
}

func TestList_ReturnsRoster(t *testing.T) {
	org := uuid.New()
	p := &domain.Principal{UserID: uuid.New(), OrgID: org, Role: domain.RoleAdmin}
	rec, fs := serve(t, p, "/api/v1/users")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, org, fs.gotOrg)

	var rows []store.UserListRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "Ann", rows[0].DisplayName)
}

func TestList_ServiceForbidden(t *testing.T) {
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleService}
	rec, _ := serve(t, p, "/api/v1/users")
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
