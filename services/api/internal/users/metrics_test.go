package users_test

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/live-rack/services/api/internal/users"
)

type fakeStats struct{ s store.RosterStats }

func (f fakeStats) RosterStatsByOrg(context.Context, uuid.UUID) (store.RosterStats, error) {
	return f.s, nil
}

type fakePending struct {
	n   int
	err error
}

func (f fakePending) PendingInvites(context.Context, string) (int, error) { return f.n, f.err }

func TestStats_ComputesCoverageAndPending(t *testing.T) {
	st := fakeStats{s: store.RosterStats{Members: 12, ActiveNow: 7, MFAUsers: 11, Roles: 5}}
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), IDPOrgID: "zid", Role: domain.RoleManager}

	e := echo.New()
	users.NewMetrics(st, fakePending{n: 2}).Register(e.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodGet, "/api/v1/users/stats", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var out users.StatsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, 12, out.Members)
	assert.Equal(t, 7, out.ActiveNow)
	assert.Equal(t, 2, out.PendingInvites)
	assert.Equal(t, 91, out.TwoFACoverage) // 11/12 -> 91%
}

func TestStats_PendingDegradesToZero(t *testing.T) {
	st := fakeStats{s: store.RosterStats{Members: 4, MFAUsers: 4, Roles: 2}}
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), IDPOrgID: "zid", Role: domain.RoleAdmin}

	e := echo.New()
	users.NewMetrics(st, fakePending{err: errors.New("zitadel down")}).Register(e.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodGet, "/api/v1/users/stats", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var out users.StatsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, 0, out.PendingInvites)
	assert.Equal(t, 100, out.TwoFACoverage)
}

func TestSyncMFA_SetsCallerFlag(t *testing.T) {
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleStaff}
	fs := &fakeStore{}
	e := echo.New()
	users.New(fs).Register(e.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodPost, "/api/v1/me/2fa",
		strings.NewReader(`{"enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.True(t, fs.mfaSet)
}
