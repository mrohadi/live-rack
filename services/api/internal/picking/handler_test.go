package picking_test

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
	"github.com/live-rack/services/api/internal/picking"
)

type source struct {
	zoneID uuid.UUID
	x, y   float64
}

type fakeStore struct {
	list      store.PickList
	lines     []store.ListPickListLinesRow
	sources   map[string]source // sku -> source location
	adjusts   []store.AdjustItemLocationQtyParams
	tasks     []store.CreateTaskParams
	knownSKUs map[string]bool
}

func (f *fakeStore) CreatePickList(_ context.Context, arg store.CreatePickListParams) (store.PickList, error) {
	f.list = store.PickList{ID: uuid.New(), OrgID: arg.OrgID, StoreID: arg.StoreID, Reference: arg.Reference, Status: "open"}
	return f.list, nil
}

func (f *fakeStore) AddPickLine(_ context.Context, arg store.AddPickLineParams) (store.PickListLine, error) {
	row := store.ListPickListLinesRow{
		ID: uuid.New(), ZoneID: arg.ZoneID, Sku: arg.Sku,
		QtyRequested: arg.QtyRequested, Seq: arg.Seq, Status: "pending",
	}
	if s, ok := f.sources[arg.Sku]; ok {
		row.ZoneX, row.ZoneY = s.x, s.y
	}
	f.lines = append(f.lines, row)
	return store.PickListLine{ID: row.ID, Sku: arg.Sku, QtyRequested: arg.QtyRequested, Seq: arg.Seq, Status: "pending"}, nil
}

func (f *fakeStore) GetPickList(_ context.Context, _ store.GetPickListParams) (store.PickList, error) {
	if f.list.ID == uuid.Nil {
		return store.PickList{}, pgx.ErrNoRows
	}
	return f.list, nil
}

func (f *fakeStore) ListPickListsByStore(_ context.Context, _ store.ListPickListsByStoreParams) ([]store.ListPickListsByStoreRow, error) {
	return nil, nil
}

func (f *fakeStore) ListPickListLines(_ context.Context, _ uuid.UUID) ([]store.ListPickListLinesRow, error) {
	return f.lines, nil
}

func (f *fakeStore) SetPickLinePicked(_ context.Context, arg store.SetPickLinePickedParams) (store.PickListLine, error) {
	for i := range f.lines {
		if f.lines[i].ID == arg.ID {
			f.lines[i].QtyPicked = arg.QtyPicked
			f.lines[i].Status = arg.Status
			r := f.lines[i]
			return store.PickListLine{ID: r.ID, Sku: r.Sku, QtyRequested: r.QtyRequested, QtyPicked: r.QtyPicked, Seq: r.Seq, Status: r.Status}, nil
		}
	}
	return store.PickListLine{}, pgx.ErrNoRows
}

func (f *fakeStore) StartPickList(_ context.Context, _ store.StartPickListParams) (store.PickList, error) {
	f.list.Status = "picking"
	return f.list, nil
}

func (f *fakeStore) CompletePickList(_ context.Context, _ store.CompletePickListParams) (store.PickList, error) {
	f.list.Status = "completed"
	return f.list, nil
}

func (f *fakeStore) ResolvePickSource(_ context.Context, arg store.ResolvePickSourceParams) (store.ResolvePickSourceRow, error) {
	s, ok := f.sources[arg.Sku]
	if !ok {
		return store.ResolvePickSourceRow{}, pgx.ErrNoRows
	}
	return store.ResolvePickSourceRow{ZoneID: s.zoneID, Qty: 100, ZoneX: s.x, ZoneY: s.y}, nil
}

func (f *fakeStore) AdjustItemLocationQty(_ context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error) {
	f.adjusts = append(f.adjusts, arg)
	return store.ItemLocation{Sku: arg.Sku, ZoneID: arg.ZoneID, Qty: 0}, nil
}

func (f *fakeStore) GetItemBySKU(_ context.Context, arg store.GetItemBySKUParams) (store.Item, error) {
	if f.knownSKUs[arg.Sku] {
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

func TestPicking_CreateOptimisesRoute(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	near, far := uuid.New(), uuid.New()
	fs := &fakeStore{sources: map[string]source{
		"FAR":  {zoneID: far, x: 100, y: 100},
		"NEAR": {zoneID: near, x: 5, y: 0},
	}}
	e := echo.New()
	e.HideBanner = true
	h := picking.New(fs, nil, nil)

	body := `{"reference":"SO-1","lines":[{"sku":"FAR","qty":2},{"sku":"NEAR","qty":1}]}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/pick-lists", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var out picking.Board
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Lines, 2)
	// nearest stop first
	assert.Equal(t, "NEAR", out.Lines[0].SKU)
	assert.EqualValues(t, 0, out.Lines[0].Seq)
	assert.Equal(t, "FAR", out.Lines[1].SKU)
	assert.EqualValues(t, 1, out.Lines[1].Seq)
}

func TestPicking_ShortPickFlagsAndRestocks(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	zone := uuid.New()
	lineID := uuid.New()
	fs := &fakeStore{
		list:      store.PickList{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "picking"},
		knownSKUs: map[string]bool{"SKU-A": true},
		lines: []store.ListPickListLinesRow{{
			ID: lineID, ZoneID: pgtype.UUID{Bytes: zone, Valid: true},
			Sku: "SKU-A", QtyRequested: 10, Seq: 0, Status: "pending",
		}},
	}
	aud := &fakeAuditor{}
	e := echo.New()
	e.HideBanner = true
	h := picking.New(fs, nil, aud)

	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/pick-lists/"+fs.list.ID.String()+"/lines/"+lineID.String(),
		strings.NewReader(`{"qty_picked":4}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id", "lineID")
	c.SetParamValues(storeID.String(), fs.list.ID.String(), lineID.String())

	require.NoError(t, h.Pick(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var out picking.LineRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, string(domain.PickLineShort), out.Status)
	assert.EqualValues(t, 4, out.QtyPicked)

	// inventory decremented by the picked qty
	require.Len(t, fs.adjusts, 1)
	assert.EqualValues(t, -4, fs.adjusts[0].Qty)
	// short pick raises a restock task
	require.Len(t, fs.tasks, 1)
	assert.Equal(t, domain.RestockTaskTitle("SKU-A"), fs.tasks[0].Title)
	require.Len(t, aud.entries, 1)
	assert.Equal(t, "inventory.pick", aud.entries[0].Action)
}

func TestPicking_FullPickNoRestock(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	zone := uuid.New()
	lineID := uuid.New()
	fs := &fakeStore{
		list:      store.PickList{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "picking"},
		knownSKUs: map[string]bool{"SKU-A": true},
		lines: []store.ListPickListLinesRow{{
			ID: lineID, ZoneID: pgtype.UUID{Bytes: zone, Valid: true},
			Sku: "SKU-A", QtyRequested: 5, Seq: 0, Status: "pending",
		}},
	}
	e := echo.New()
	e.HideBanner = true
	h := picking.New(fs, nil, nil)

	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/pick-lists/"+fs.list.ID.String()+"/lines/"+lineID.String(),
		strings.NewReader(`{"qty_picked":5}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id", "lineID")
	c.SetParamValues(storeID.String(), fs.list.ID.String(), lineID.String())

	require.NoError(t, h.Pick(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var out picking.LineRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, string(domain.PickLinePicked), out.Status)
	assert.Empty(t, fs.tasks)
	require.Len(t, fs.adjusts, 1)
	assert.EqualValues(t, -5, fs.adjusts[0].Qty)
}

func TestPicking_PickRejectsOverpick(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	lineID := uuid.New()
	fs := &fakeStore{
		list: store.PickList{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "picking"},
		lines: []store.ListPickListLinesRow{{
			ID: lineID, Sku: "SKU-A", QtyRequested: 3, Seq: 0, Status: "pending",
		}},
	}
	e := echo.New()
	e.HideBanner = true
	h := picking.New(fs, nil, nil)

	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/pick-lists/"+fs.list.ID.String()+"/lines/"+lineID.String(),
		strings.NewReader(`{"qty_picked":9}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id", "lineID")
	c.SetParamValues(storeID.String(), fs.list.ID.String(), lineID.String())

	err := h.Pick(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}
