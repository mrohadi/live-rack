package waves_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/waves"
)

type fakeStore struct {
	wave    store.Wave
	merged  []store.ListWaveMergedLinesRow
	members []store.ListWaveStopMemberLinesRow
	setArgs []store.SetPickLinePickedParams
	adjusts []store.AdjustItemLocationQtyParams
	tasks   []store.CreateTaskParams
	known   map[string]bool
}

func (f *fakeStore) CreateWave(_ context.Context, arg store.CreateWaveParams) (store.Wave, error) {
	f.wave = store.Wave{ID: uuid.New(), OrgID: arg.OrgID, StoreID: arg.StoreID, Reference: arg.Reference, Status: "open"}
	return f.wave, nil
}
func (f *fakeStore) AssignListsToWave(_ context.Context, _ store.AssignListsToWaveParams) error {
	return nil
}
func (f *fakeStore) GetWave(_ context.Context, _ store.GetWaveParams) (store.Wave, error) {
	if f.wave.ID == uuid.Nil {
		return store.Wave{}, pgx.ErrNoRows
	}
	return f.wave, nil
}
func (f *fakeStore) ListWavesByStore(_ context.Context, _ store.ListWavesByStoreParams) ([]store.ListWavesByStoreRow, error) {
	return nil, nil
}
func (f *fakeStore) ListWaveMergedLines(_ context.Context, _ store.ListWaveMergedLinesParams) ([]store.ListWaveMergedLinesRow, error) {
	return f.merged, nil
}
func (f *fakeStore) ListWaveStopMemberLines(_ context.Context, _ store.ListWaveStopMemberLinesParams) ([]store.ListWaveStopMemberLinesRow, error) {
	return f.members, nil
}
func (f *fakeStore) StartWave(_ context.Context, _ store.StartWaveParams) (store.Wave, error) {
	f.wave.Status = "picking"
	return f.wave, nil
}
func (f *fakeStore) CompleteWave(_ context.Context, _ store.CompleteWaveParams) (store.Wave, error) {
	f.wave.Status = "completed"
	return f.wave, nil
}
func (f *fakeStore) SetPickLinePicked(_ context.Context, arg store.SetPickLinePickedParams) (store.PickListLine, error) {
	f.setArgs = append(f.setArgs, arg)
	return store.PickListLine{ID: arg.ID, QtyPicked: arg.QtyPicked, Status: arg.Status}, nil
}
func (f *fakeStore) AdjustItemLocationQty(_ context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error) {
	f.adjusts = append(f.adjusts, arg)
	return store.ItemLocation{Sku: arg.Sku, Qty: 0}, nil
}
func (f *fakeStore) GetItemBySKU(_ context.Context, arg store.GetItemBySKUParams) (store.Item, error) {
	if f.known[arg.Sku] {
		return store.Item{Sku: arg.Sku}, nil
	}
	return store.Item{}, pgx.ErrNoRows
}
func (f *fakeStore) CountOpenTasksByTitle(_ context.Context, _ store.CountOpenTasksByTitleParams) (int64, error) {
	return 0, nil
}
func (f *fakeStore) CreateTask(_ context.Context, arg store.CreateTaskParams) (store.Task, error) {
	f.tasks = append(f.tasks, arg)
	return store.Task{ID: uuid.New()}, nil
}

type fakeAuditor struct{ entries []audit.Entry }

func (a *fakeAuditor) Write(_ context.Context, e audit.Entry) error {
	a.entries = append(a.entries, e)
	return nil
}

func ctxFor(orgID uuid.UUID) context.Context {
	return pkgauth.WithPrincipal(context.Background(),
		&domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff})
}

func TestWaves_CreateMergesAndOrdersRoute(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	near, far := uuid.New(), uuid.New()
	fs := &fakeStore{merged: []store.ListWaveMergedLinesRow{
		{Sku: "FAR", ZoneID: pgtype.UUID{Bytes: far, Valid: true}, ZoneX: 100, ZoneY: 100, QtyRequested: 8, OrderCount: 2},
		{Sku: "NEAR", ZoneID: pgtype.UUID{Bytes: near, Valid: true}, ZoneX: 5, ZoneY: 0, QtyRequested: 3, OrderCount: 2},
	}}
	e := echo.New()
	e.HideBanner = true
	h := waves.New(fs, nil, nil)

	l1, l2 := uuid.New(), uuid.New()
	body := `{"reference":"WAVE-1","list_ids":["` + l1.String() + `","` + l2.String() + `"]}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/waves", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var out waves.Board
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Stops, 2)
	assert.Equal(t, "NEAR", out.Stops[0].SKU)
	assert.EqualValues(t, 0, out.Stops[0].Seq)
	assert.Equal(t, "FAR", out.Stops[1].SKU)
	assert.EqualValues(t, 2, out.Stops[0].OrderCount)
}

func TestWaves_CreateNeedsTwoLists(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	e.HideBanner = true
	h := waves.New(fs, nil, nil)

	body := `{"list_ids":["` + uuid.New().String() + `"]}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/waves", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestWaves_PickStopAllocatesFIFOAndRestocks(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	zone := uuid.New()
	la, lb := uuid.New(), uuid.New()
	fs := &fakeStore{
		wave:  store.Wave{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "picking"},
		known: map[string]bool{"SKU-A": true},
		members: []store.ListWaveStopMemberLinesRow{
			{ID: la, QtyRequested: 5},
			{ID: lb, QtyRequested: 4},
		},
		merged: []store.ListWaveMergedLinesRow{
			{Sku: "SKU-A", ZoneID: pgtype.UUID{Bytes: zone, Valid: true}, QtyRequested: 9, QtyPicked: 7, OrderCount: 2},
		},
	}
	aud := &fakeAuditor{}
	e := echo.New()
	e.HideBanner = true
	h := waves.New(fs, nil, aud)

	body := `{"sku":"SKU-A","zone_id":"` + zone.String() + `","qty_picked":7}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/waves/"+fs.wave.ID.String()+"/stops",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), fs.wave.ID.String())

	require.NoError(t, h.PickStop(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	// FIFO: A filled to 5 (picked), B gets 2 (short).
	require.Len(t, fs.setArgs, 2)
	assert.Equal(t, la, fs.setArgs[0].ID)
	assert.EqualValues(t, 5, fs.setArgs[0].QtyPicked)
	assert.Equal(t, string(domain.PickLinePicked), fs.setArgs[0].Status)
	assert.Equal(t, lb, fs.setArgs[1].ID)
	assert.EqualValues(t, 2, fs.setArgs[1].QtyPicked)
	assert.Equal(t, string(domain.PickLineShort), fs.setArgs[1].Status)

	// inventory decremented once by the merged qty
	require.Len(t, fs.adjusts, 1)
	assert.EqualValues(t, -7, fs.adjusts[0].Qty)
	// short overall (7 < 9) → restock raised
	require.Len(t, fs.tasks, 1)
	require.Len(t, aud.entries, 1)
	assert.Equal(t, "inventory.wave_pick", aud.entries[0].Action)
}
