package inventory_test

import (
	"bytes"
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
	"github.com/live-rack/services/api/internal/inventory"
)

// fakeStore satisfies inventory.Store.
type fakeStore struct {
	// list
	gotOrg   uuid.UUID
	gotStore uuid.UUID
	rows     []store.ListInventoryByStoreRow

	// write
	upsertArg    store.UpsertItemParams
	adjustArg    store.AdjustItemLocationQtyParams
	adjustResult store.ItemLocation
}

func (f *fakeStore) ListInventoryByStore(_ context.Context, arg store.ListInventoryByStoreParams) ([]store.ListInventoryByStoreRow, error) {
	f.gotOrg = arg.OrgID
	f.gotStore = arg.StoreID
	return f.rows, nil
}

func (f *fakeStore) UpsertItem(_ context.Context, arg store.UpsertItemParams) (store.Item, error) {
	f.upsertArg = arg
	return store.Item{OrgID: arg.OrgID, Sku: arg.Sku, Name: arg.Name, Category: arg.Category, Status: arg.Status}, nil
}

func (f *fakeStore) AdjustItemLocationQty(_ context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error) {
	f.adjustArg = arg
	f.adjustResult.ID = uuid.New()
	f.adjustResult.OrgID = arg.OrgID
	f.adjustResult.StoreID = arg.StoreID
	f.adjustResult.ZoneID = arg.ZoneID
	f.adjustResult.Sku = arg.Sku
	f.adjustResult.Qty = arg.Qty
	f.adjustResult.UpdatedAt = time.Now().UTC()
	return f.adjustResult, nil
}

func newCtx(orgID uuid.UUID) context.Context {
	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleAdmin}
	return pkgauth.WithPrincipal(context.Background(), p)
}

func TestInventoryHandler_List(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	zoneID := uuid.New()

	fs := &fakeStore{rows: []store.ListInventoryByStoreRow{{
		ID: uuid.New(), OrgID: orgID, StoreID: storeID, ZoneID: zoneID,
		Sku: "SKU-1", Qty: 7, Name: "Widget", Category: "frozen", Status: "active",
		Picks7d: 8, Picks30d: 20,
		UpdatedAt: time.Now().UTC(),
	}}}

	e := echo.New()
	e.HideBanner = true
	h := inventory.New(fs)
	h.Register(e.Group("/api/v1/stores"))

	req := httptest.NewRequestWithContext(newCtx(orgID), http.MethodGet,
		"/api/v1/stores/"+storeID.String()+"/inventory", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var out []inventory.Row
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "SKU-1", out[0].SKU)
	assert.EqualValues(t, 7, out[0].Qty)
	assert.Equal(t, "hot", out[0].Velocity)
}

func TestInventoryHandler_Add(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	zoneID := uuid.New()

	fs := &fakeStore{}

	e := echo.New()
	e.HideBanner = true
	h := inventory.New(fs)

	body := inventory.AddRequest{
		ZoneID:   zoneID.String(),
		SKU:      "SKU-42",
		Name:     "Test Widget",
		Category: "general",
		Qty:      5,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequestWithContext(newCtx(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/inventory",
		bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Add(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify item upserted with correct fields.
	assert.Equal(t, orgID, fs.upsertArg.OrgID)
	assert.Equal(t, "SKU-42", fs.upsertArg.Sku)
	assert.Equal(t, "Test Widget", fs.upsertArg.Name)
	assert.Equal(t, "active", fs.upsertArg.Status) // defaulted

	// Verify qty adjusted.
	assert.Equal(t, orgID, fs.adjustArg.OrgID)
	assert.Equal(t, storeID, fs.adjustArg.StoreID)
	assert.Equal(t, zoneID, fs.adjustArg.ZoneID)
	assert.EqualValues(t, 5, fs.adjustArg.Qty)

	var out inventory.Row
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "SKU-42", out.SKU)
	assert.EqualValues(t, 5, out.Qty)
	assert.Equal(t, "cold", out.Velocity)
}

func TestInventoryHandler_Add_Validation(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	zoneID := uuid.New()

	cases := []struct {
		name   string
		body   string
		status int
	}{
		{"missing sku", `{"zone_id":"` + zoneID.String() + `","qty":1}`, http.StatusBadRequest},
		{"qty zero", `{"zone_id":"` + zoneID.String() + `","sku":"X","qty":0}`, http.StatusBadRequest},
		{"bad zone_id", `{"zone_id":"not-uuid","sku":"X","qty":1}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := &fakeStore{}
			e := echo.New()
			e.HideBanner = true
			h := inventory.New(fs)

			req := httptest.NewRequestWithContext(newCtx(orgID), http.MethodPost,
				"/api/v1/stores/"+storeID.String()+"/inventory",
				strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("storeID")
			c.SetParamValues(storeID.String())

			err := h.Add(c)
			require.Error(t, err)
			he, ok := err.(*echo.HTTPError)
			require.True(t, ok)
			assert.Equal(t, tc.status, he.Code)
		})
	}
}
