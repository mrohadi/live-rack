package search_test

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
	"github.com/live-rack/services/api/internal/search"
)

type fakeSearcher struct {
	gotArg store.SearchEntitiesParams
	rows   []store.SearchEntitiesRow
}

func (f *fakeSearcher) SearchEntities(_ context.Context, arg store.SearchEntitiesParams) ([]store.SearchEntitiesRow, error) {
	f.gotArg = arg
	return f.rows, nil
}

func doSearch(t *testing.T, s *fakeSearcher, orgID uuid.UUID, target string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.HideBanner = true
	search.New(s).Register(e.Group("/api/v1"))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleAdmin}
	req = req.WithContext(pkgauth.WithPrincipal(req.Context(), p))
	e.ServeHTTP(rec, req)
	return rec
}

func TestSearch_ReturnsHits(t *testing.T) {
	orgID := uuid.New()
	s := &fakeSearcher{rows: []store.SearchEntitiesRow{
		{Kind: "item", ID: uuid.New(), Label: "Widget", Sublabel: "SKU-1", Score: 0.9},
		{Kind: "zone", ID: uuid.New(), Label: "Frozen Bay", Sublabel: "frozen", Score: 0.4},
	}}

	rec := doSearch(t, s, orgID, "/api/v1/search?q=wid&limit=5")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "wid", s.gotArg.Query)
	assert.Equal(t, orgID, s.gotArg.OrgID)
	assert.EqualValues(t, 5, s.gotArg.MaxResults)

	var out []search.Result
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 2)
	assert.Equal(t, "item", out[0].Kind)
	assert.Equal(t, "Widget", out[0].Label)
}

func TestSearch_RejectsShortQuery(t *testing.T) {
	rec := doSearch(t, &fakeSearcher{}, uuid.New(), "/api/v1/search?q=a")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSearch_DefaultsLimit(t *testing.T) {
	s := &fakeSearcher{}
	doSearch(t, s, uuid.New(), "/api/v1/search?q=widget")
	assert.EqualValues(t, 20, s.gotArg.MaxResults)
}

func TestSearch_ClampsLimit(t *testing.T) {
	s := &fakeSearcher{}
	doSearch(t, s, uuid.New(), "/api/v1/search?q=widget&limit=999")
	assert.EqualValues(t, 50, s.gotArg.MaxResults)
}
