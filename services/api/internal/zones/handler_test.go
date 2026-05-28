package zones_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/zones"
)

// fakeZoneStore implements zones.ZoneStore for unit tests.
type fakeZoneStore struct {
	rows map[uuid.UUID]store.Zone
}

func newFakeStore() *fakeZoneStore {
	return &fakeZoneStore{rows: map[uuid.UUID]store.Zone{}}
}

func (f *fakeZoneStore) CreateZone(_ context.Context, arg store.CreateZoneParams) (store.Zone, error) {
	z := store.Zone{
		ID:          uuid.New(),
		OrgID:       arg.OrgID,
		StoreID:     arg.StoreID,
		Name:        arg.Name,
		Type:        arg.Type,
		X:           arg.X,
		Y:           arg.Y,
		Width:       arg.Width,
		Height:      arg.Height,
		Color:       arg.Color,
		Capacity:    arg.Capacity,
		Constraints: arg.Constraints,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	f.rows[z.ID] = z
	return z, nil
}

func (f *fakeZoneStore) GetZone(_ context.Context, arg store.GetZoneParams) (store.Zone, error) {
	z, ok := f.rows[arg.ID]
	if !ok || z.OrgID != arg.OrgID {
		return store.Zone{}, &notFoundError{id: arg.ID}
	}
	return z, nil
}

func (f *fakeZoneStore) ListZonesByStore(_ context.Context, arg store.ListZonesByStoreParams) ([]store.Zone, error) {
	var out []store.Zone
	for _, z := range f.rows {
		if z.StoreID == arg.StoreID && z.OrgID == arg.OrgID {
			out = append(out, z)
		}
	}
	return out, nil
}

func (f *fakeZoneStore) UpdateZone(_ context.Context, arg store.UpdateZoneParams) (store.Zone, error) {
	z, ok := f.rows[arg.ID]
	if !ok || z.OrgID != arg.OrgID {
		return store.Zone{}, &notFoundError{id: arg.ID}
	}
	z.Name = arg.Name
	z.Type = arg.Type
	z.X = arg.X
	z.Y = arg.Y
	z.Width = arg.Width
	z.Height = arg.Height
	z.Color = arg.Color
	z.Capacity = arg.Capacity
	z.Constraints = arg.Constraints
	z.UpdatedAt = time.Now()
	f.rows[z.ID] = z
	return z, nil
}

func (f *fakeZoneStore) DeleteZone(_ context.Context, arg store.DeleteZoneParams) error {
	z, ok := f.rows[arg.ID]
	if !ok || z.OrgID != arg.OrgID {
		return &notFoundError{id: arg.ID}
	}
	delete(f.rows, arg.ID)
	return nil
}

type notFoundError struct{ id uuid.UUID }

func (e *notFoundError) Error() string { return "zone not found: " + e.id.String() }

// withPrincipal injects a Principal into the Echo request context.
func withPrincipal(c echo.Context, orgID, storeID uuid.UUID) {
	p := &domain.Principal{
		UserID:     uuid.New(),
		OrgID:      orgID,
		ClerkOrgID: "clerk_" + orgID.String(),
		Role:       domain.RoleAdmin,
		StoreIDs:   []uuid.UUID{storeID},
	}
	ctx := pkgauth.WithPrincipal(c.Request().Context(), p)
	c.SetRequest(c.Request().WithContext(ctx))
}

func newEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	return e
}

func TestZoneHandler_Create(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	body, _ := json.Marshal(map[string]any{
		"name":        "Zone A",
		"type":        "general",
		"x":           10.0,
		"y":           20.0,
		"width":       100.0,
		"height":      80.0,
		"color":       "#6366f1",
		"capacity":    50,
		"constraints": map[string]any{},
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/zones", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	withPrincipal(c, orgID, storeID)

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Zone A", resp["name"])
	assert.Equal(t, "general", resp["type"])
	assert.NotEmpty(t, resp["id"])
}

func TestZoneHandler_Create_BadRequest(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	// Missing required name.
	body, _ := json.Marshal(map[string]any{
		"type": "general",
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/zones", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	withPrincipal(c, orgID, storeID)

	err := h.Create(c)
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestZoneHandler_Create_RejectsInvalidConstraints(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	// Allowed and denied overlap — Validate() must reject.
	body, _ := json.Marshal(map[string]any{
		"name":     "Bad Zone",
		"type":     "general",
		"x":        0.0,
		"y":        0.0,
		"width":    50.0,
		"height":   50.0,
		"color":    "#fff",
		"capacity": 10,
		"constraints": map[string]any{
			"allowed_categories": []string{"frozen"},
			"denied_categories":  []string{"frozen"},
		},
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/zones", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	withPrincipal(c, orgID, storeID)

	err := h.Create(c)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok, "expected echo.HTTPError, got %T", err)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Empty(t, fs.rows, "no zone should be persisted on validation failure")
}

func TestZoneHandler_Create_PersistsValidConstraints(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	body, _ := json.Marshal(map[string]any{
		"name": "Frozen Aisle",
		"type": "frozen",
		"x":    0.0, "y": 0.0, "width": 50.0, "height": 50.0,
		"color":    "#0ea5e9",
		"capacity": 100,
		"constraints": map[string]any{
			"allowed_categories": []string{"frozen", "dairy"},
			"max_units_per_sku":  25,
			"require_dual_scan":  true,
		},
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/zones", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	withPrincipal(c, orgID, storeID)

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	require.Len(t, fs.rows, 1)
	var stored store.Zone
	for _, z := range fs.rows {
		stored = z
	}

	parsed, err := domain.UnmarshalConstraints(stored.Constraints)
	require.NoError(t, err)
	assert.Equal(t, []string{"frozen", "dairy"}, parsed.AllowedCategories)
	require.NotNil(t, parsed.MaxUnitsPerSKU)
	assert.Equal(t, 25, *parsed.MaxUnitsPerSKU)
	assert.True(t, parsed.RequireDualScan)
}

func TestZoneHandler_Update_RejectsInvalidConstraints(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	ctx := context.Background()
	created, err := fs.CreateZone(ctx, store.CreateZoneParams{
		OrgID: orgID, StoreID: storeID,
		Name: "Zone", Type: store.ZoneTypeGeneral,
		X: 0, Y: 0, Width: 50, Height: 50,
		Color: "#fff", Capacity: 10, Constraints: []byte(`{}`),
	})
	require.NoError(t, err)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	body, _ := json.Marshal(map[string]any{
		"name": "Zone",
		"type": "general",
		"x":    0.0, "y": 0.0, "width": 50.0, "height": 50.0,
		"color":    "#fff",
		"capacity": 10,
		"constraints": map[string]any{
			"max_units_per_sku": -3,
		},
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut,
		"/api/v1/stores/"+storeID.String()+"/zones/"+created.ID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), created.ID.String())
	withPrincipal(c, orgID, storeID)

	err = h.Update(c)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestZoneHandler_Get(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	ctx := context.Background()
	created, err := fs.CreateZone(ctx, store.CreateZoneParams{
		OrgID: orgID, StoreID: storeID,
		Name: "Zone B", Type: store.ZoneTypeGeneral,
		X: 0, Y: 0, Width: 50, Height: 50,
		Color: "#fff", Capacity: 10, Constraints: []byte(`{}`),
	})
	require.NoError(t, err)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/stores/"+storeID.String()+"/zones/"+created.ID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), created.ID.String())
	withPrincipal(c, orgID, storeID)

	require.NoError(t, h.Get(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, created.ID.String(), resp["id"])
	assert.Equal(t, "Zone B", resp["name"])
}

func TestZoneHandler_List(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	ctx := context.Background()
	for i := range 3 {
		_, err := fs.CreateZone(ctx, store.CreateZoneParams{
			OrgID: orgID, StoreID: storeID,
			Name: "Zone " + string(rune('A'+i)), Type: store.ZoneTypeGeneral,
			X: 0, Y: 0, Width: 50, Height: 50,
			Color: "#fff", Capacity: 0, Constraints: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/stores/"+storeID.String()+"/zones", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	withPrincipal(c, orgID, storeID)

	require.NoError(t, h.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 3)
}

func TestZoneHandler_Update(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	ctx := context.Background()
	created, err := fs.CreateZone(ctx, store.CreateZoneParams{
		OrgID: orgID, StoreID: storeID,
		Name: "Zone Old", Type: store.ZoneTypeGeneral,
		X: 0, Y: 0, Width: 50, Height: 50,
		Color: "#fff", Capacity: 10, Constraints: []byte(`{}`),
	})
	require.NoError(t, err)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	body, _ := json.Marshal(map[string]any{
		"name":        "Zone New",
		"type":        "frozen",
		"x":           5.0,
		"y":           5.0,
		"width":       60.0,
		"height":      60.0,
		"color":       "#0ea5e9",
		"capacity":    20,
		"constraints": map[string]any{},
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut,
		"/api/v1/stores/"+storeID.String()+"/zones/"+created.ID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), created.ID.String())
	withPrincipal(c, orgID, storeID)

	require.NoError(t, h.Update(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Zone New", resp["name"])
	assert.Equal(t, "frozen", resp["type"])
}

func TestZoneHandler_Delete(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	ctx := context.Background()
	created, err := fs.CreateZone(ctx, store.CreateZoneParams{
		OrgID: orgID, StoreID: storeID,
		Name: "Zone Del", Type: store.ZoneTypeGeneral,
		X: 0, Y: 0, Width: 50, Height: 50,
		Color: "#fff", Capacity: 0, Constraints: []byte(`{}`),
	})
	require.NoError(t, err)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete,
		"/api/v1/stores/"+storeID.String()+"/zones/"+created.ID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), created.ID.String())
	withPrincipal(c, orgID, storeID)

	require.NoError(t, h.Delete(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestZoneHandler_Get_NotFound(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	fs := newFakeStore()
	h := zones.New(fs)

	e := newEcho()
	h.Register(e.Group("/api/v1/stores"))

	missingID := uuid.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/stores/"+storeID.String()+"/zones/"+missingID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), missingID.String())
	withPrincipal(c, orgID, storeID)

	err := h.Get(c)
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusNotFound, he.Code)
}
