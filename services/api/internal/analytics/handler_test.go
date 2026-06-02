package analytics_test

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
	"github.com/live-rack/services/api/internal/analytics"
)

type fakeReader struct {
	body   []byte
	gotSQL string
}

func (f *fakeReader) Query(_ context.Context, sql string) ([]byte, error) {
	f.gotSQL = sql
	return f.body, nil
}

func serve(t *testing.T, fr *fakeReader, org uuid.UUID, target string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	analytics.New(fr).Register(e.Group("/api/v1"))
	p := &domain.Principal{UserID: uuid.New(), OrgID: org, Role: domain.RoleReadonly}
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestHeatmap_FoldsGrid(t *testing.T) {
	org := uuid.New()
	fr := &fakeReader{body: []byte(`{"data":[
		{"dow":1,"hour":9,"scans":5},
		{"dow":7,"hour":23,"scans":12}
	]}`)}

	rec := serve(t, fr, org, "/api/v1/analytics/heatmap")
	require.Equal(t, http.StatusOK, rec.Code)

	var out analytics.HeatmapResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Grid, 7)
	require.Len(t, out.Grid[0], 24)
	assert.Equal(t, int64(5), out.Grid[0][9])   // Monday 09:00
	assert.Equal(t, int64(12), out.Grid[6][23]) // Sunday 23:00
	assert.Equal(t, int64(12), out.Max)
	assert.Contains(t, fr.gotSQL, "org_id = '"+org.String()+"'")
}

func TestHeatmap_ZoneFilter(t *testing.T) {
	org, zone := uuid.New(), uuid.New()
	fr := &fakeReader{body: []byte(`{"data":[]}`)}
	rec := serve(t, fr, org, "/api/v1/analytics/heatmap?zone_id="+zone.String())
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, fr.gotSQL, "zone_id = '"+zone.String()+"'")
}

func TestHeatmap_BadZone(t *testing.T) {
	fr := &fakeReader{body: []byte(`{"data":[]}`)}
	rec := serve(t, fr, uuid.New(), "/api/v1/analytics/heatmap?zone_id=not-a-uuid")
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// seqReader returns a different body per Query call, in order.
type seqReader struct {
	bodies [][]byte
	n      int
}

func (s *seqReader) Query(_ context.Context, _ string) ([]byte, error) {
	b := s.bodies[s.n]
	if s.n < len(s.bodies)-1 {
		s.n++
	}
	return b, nil
}

func TestZones_JoinsTotalsAndSpark(t *testing.T) {
	org := uuid.New()
	za, zb := uuid.New().String(), uuid.New().String()
	sr := &seqReader{bodies: [][]byte{
		[]byte(`{"data":[{"zone_id":"` + za + `","scans":30,"picks":20,"invalid":2},{"zone_id":"` + zb + `","scans":5,"picks":1,"invalid":0}]}`),
		[]byte(`{"data":[{"zone_id":"` + za + `","scans":10},{"zone_id":"` + za + `","scans":20},{"zone_id":"` + zb + `","scans":5}]}`),
	}}

	e := echo.New()
	analytics.New(sr).Register(e.Group("/api/v1"))
	p := &domain.Principal{UserID: uuid.New(), OrgID: org, Role: domain.RoleReadonly}
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodGet, "/api/v1/analytics/zones", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out analytics.ZonesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Zones, 2)
	assert.Equal(t, za, out.Zones[0].ZoneID) // ordered by scans desc
	assert.Equal(t, int64(30), out.Zones[0].Scans)
	assert.Equal(t, []int64{10, 20}, out.Zones[0].Spark)
	assert.Equal(t, []int64{5}, out.Zones[1].Spark)
}
