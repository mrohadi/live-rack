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

	"github.com/jackc/pgx/v5"

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

	// transfer
	decrementArg   store.DecrementItemLocationQtyParams
	decrementErr   error
	decrementCalls int

	// detail
	item     store.Item
	itemErr  error
	skuLocs  []store.ListItemLocationsBySKURow
	skuScans []store.ScanEvent

	// edit / adjust
	updateArg store.UpdateItemParams
	updateErr error
	setArg    store.SetItemLocationQtyParams
	setErr    error
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

func (f *fakeStore) DecrementItemLocationQty(_ context.Context, arg store.DecrementItemLocationQtyParams) (store.ItemLocation, error) {
	f.decrementArg = arg
	f.decrementCalls++
	if f.decrementErr != nil {
		return store.ItemLocation{}, f.decrementErr
	}
	return store.ItemLocation{ID: uuid.New(), OrgID: arg.OrgID, ZoneID: arg.ZoneID, Sku: arg.Sku}, nil
}

func (f *fakeStore) GetItemBySKU(_ context.Context, arg store.GetItemBySKUParams) (store.Item, error) {
	if f.itemErr != nil {
		return store.Item{}, f.itemErr
	}
	if f.item.Sku == "" {
		f.item.Sku = arg.Sku
	}
	return f.item, nil
}

func (f *fakeStore) ListItemLocationsBySKU(_ context.Context, _ store.ListItemLocationsBySKUParams) ([]store.ListItemLocationsBySKURow, error) {
	return f.skuLocs, nil
}

func (f *fakeStore) ListScanEventsBySKU(_ context.Context, _ store.ListScanEventsBySKUParams) ([]store.ScanEvent, error) {
	return f.skuScans, nil
}

func (f *fakeStore) UpdateItem(_ context.Context, arg store.UpdateItemParams) (store.Item, error) {
	f.updateArg = arg
	if f.updateErr != nil {
		return store.Item{}, f.updateErr
	}
	return store.Item{
		OrgID: arg.OrgID, Sku: arg.Sku, Name: arg.Name,
		Category: arg.Category, Status: arg.Status, ReorderPoint: arg.ReorderPoint,
	}, nil
}

func (f *fakeStore) SetItemLocationQty(_ context.Context, arg store.SetItemLocationQtyParams) (store.ItemLocation, error) {
	f.setArg = arg
	if f.setErr != nil {
		return store.ItemLocation{}, f.setErr
	}
	return store.ItemLocation{ID: uuid.New(), OrgID: arg.OrgID, ZoneID: arg.ZoneID, Sku: arg.Sku, Qty: arg.Qty}, nil
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

func doTransfer(t *testing.T, fs *fakeStore, orgID, storeID uuid.UUID, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.HideBanner = true
	h := inventory.New(fs)
	req := httptest.NewRequestWithContext(newCtx(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/inventory/transfer",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	if err := h.Transfer(c); err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			require.NoError(t, c.JSON(he.Code, map[string]string{"message": he.Message.(string)}))
		}
	}
	return rec
}

func TestInventoryHandler_Transfer(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	from := uuid.New()
	to := uuid.New()

	t.Run("moves stock from source to destination", func(t *testing.T) {
		fs := &fakeStore{}
		body := `{"sku":"SKU-7","from_zone_id":"` + from.String() +
			`","to_zone_id":"` + to.String() + `","qty":4}`
		rec := doTransfer(t, fs, orgID, storeID, body)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 1, fs.decrementCalls)
		assert.Equal(t, from, fs.decrementArg.ZoneID)
		assert.EqualValues(t, 4, fs.decrementArg.Qty)
		assert.Equal(t, to, fs.adjustArg.ZoneID)
		assert.EqualValues(t, 4, fs.adjustArg.Qty)
	})

	t.Run("409 when source stock insufficient", func(t *testing.T) {
		fs := &fakeStore{decrementErr: pgx.ErrNoRows}
		body := `{"sku":"SKU-7","from_zone_id":"` + from.String() +
			`","to_zone_id":"` + to.String() + `","qty":99}`
		rec := doTransfer(t, fs, orgID, storeID, body)

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Equal(t, 1, fs.decrementCalls)
		// destination must not be credited on a failed decrement
		assert.Equal(t, uuid.Nil, fs.adjustArg.ZoneID)
	})
}

func TestInventoryHandler_Transfer_Validation(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	z := uuid.New()

	cases := []struct {
		name   string
		body   string
		status int
	}{
		{"missing sku", `{"from_zone_id":"` + z.String() + `","to_zone_id":"` + uuid.New().String() + `","qty":1}`, http.StatusBadRequest},
		{"qty zero", `{"sku":"X","from_zone_id":"` + z.String() + `","to_zone_id":"` + uuid.New().String() + `","qty":0}`, http.StatusBadRequest},
		{"same zone", `{"sku":"X","from_zone_id":"` + z.String() + `","to_zone_id":"` + z.String() + `","qty":1}`, http.StatusBadRequest},
		{"bad from zone", `{"sku":"X","from_zone_id":"nope","to_zone_id":"` + uuid.New().String() + `","qty":1}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := &fakeStore{}
			rec := doTransfer(t, fs, orgID, storeID, tc.body)
			assert.Equal(t, tc.status, rec.Code)
			assert.Equal(t, 0, fs.decrementCalls)
		})
	}
}

func TestInventoryHandler_Detail(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	z1 := uuid.New()
	z2 := uuid.New()

	t.Run("aggregates locations + scans for a sku", func(t *testing.T) {
		fs := &fakeStore{
			item: store.Item{Sku: "SKU-7", Name: "Widget", Category: "frozen", Status: "active", ReorderPoint: 5},
			skuLocs: []store.ListItemLocationsBySKURow{
				{ZoneID: z1, ZoneName: "Frozen", Qty: 3, UpdatedAt: time.Now().UTC()},
				{ZoneID: z2, ZoneName: "Backroom", Qty: 9, UpdatedAt: time.Now().UTC()},
			},
			skuScans: []store.ScanEvent{
				{ID: uuid.New(), Ts: time.Now().UTC(), ZoneID: z1, ScannerID: "scn-1", Action: "pick", Valid: true},
			},
		}
		e := echo.New()
		e.HideBanner = true
		h := inventory.New(fs)
		req := httptest.NewRequestWithContext(newCtx(orgID), http.MethodGet,
			"/api/v1/stores/"+storeID.String()+"/inventory/SKU-7", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("storeID", "sku")
		c.SetParamValues(storeID.String(), "SKU-7")

		require.NoError(t, h.Detail(c))
		assert.Equal(t, http.StatusOK, rec.Code)

		var out inventory.DetailResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
		assert.Equal(t, "SKU-7", out.SKU)
		assert.EqualValues(t, 12, out.TotalQty)
		assert.Equal(t, "in_stock", out.StockStatus) // 12 > reorder 5
		require.Len(t, out.Locations, 2)
		assert.Equal(t, "low", out.Locations[0].StockStatus) // zone1 qty 3 <= 5
		require.Len(t, out.RecentScans, 1)
		assert.Equal(t, "pick", out.RecentScans[0].Action)
	})

	t.Run("404 when item unknown", func(t *testing.T) {
		fs := &fakeStore{itemErr: pgx.ErrNoRows}
		e := echo.New()
		e.HideBanner = true
		h := inventory.New(fs)
		req := httptest.NewRequestWithContext(newCtx(orgID), http.MethodGet,
			"/api/v1/stores/"+storeID.String()+"/inventory/NOPE", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("storeID", "sku")
		c.SetParamValues(storeID.String(), "NOPE")

		err := h.Detail(c)
		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, he.Code)
	})
}
