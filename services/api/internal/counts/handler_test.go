package counts_test

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

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/counts"
)

type fakeStore struct {
	count     store.CycleCount
	lines     []store.CycleCountLine
	setQtyArg []store.SetItemLocationQtyParams
	completed bool
}

func (f *fakeStore) CreateCycleCount(_ context.Context, arg store.CreateCycleCountParams) (store.CycleCount, error) {
	f.count = store.CycleCount{
		ID: uuid.New(), OrgID: arg.OrgID, StoreID: arg.StoreID, ZoneID: arg.ZoneID, Status: "open",
	}
	return f.count, nil
}

func (f *fakeStore) SnapshotCountLines(_ context.Context, _ store.SnapshotCountLinesParams) error {
	return nil
}

func (f *fakeStore) GetCycleCount(_ context.Context, arg store.GetCycleCountParams) (store.CycleCount, error) {
	if f.count.ID == uuid.Nil {
		return store.CycleCount{}, errNotFound
	}
	return f.count, nil
}

func (f *fakeStore) ListCountLines(_ context.Context, _ uuid.UUID) ([]store.CycleCountLine, error) {
	return f.lines, nil
}

func (f *fakeStore) SetCountedQty(_ context.Context, arg store.SetCountedQtyParams) (store.CycleCountLine, error) {
	for i := range f.lines {
		if f.lines[i].Sku == arg.Sku {
			f.lines[i].CountedQty = pgtype.Int4{Int32: arg.CountedQty, Valid: true}
			return f.lines[i], nil
		}
	}
	return store.CycleCountLine{}, errNotFound
}

func (f *fakeStore) CompleteCycleCount(_ context.Context, _ store.CompleteCycleCountParams) (store.CycleCount, error) {
	f.completed = true
	f.count.Status = "completed"
	return f.count, nil
}

func (f *fakeStore) SetItemLocationQty(_ context.Context, arg store.SetItemLocationQtyParams) (store.ItemLocation, error) {
	f.setQtyArg = append(f.setQtyArg, arg)
	return store.ItemLocation{Sku: arg.Sku, ZoneID: arg.ZoneID, Qty: arg.Qty}, nil
}

type fakeAuditor struct{ entries []audit.Entry }

func (a *fakeAuditor) Write(_ context.Context, e audit.Entry) error {
	a.entries = append(a.entries, e)
	return nil
}

var errNotFound = &notFound{}

type notFound struct{}

func (e *notFound) Error() string { return "not found" }

func ctxFor(orgID uuid.UUID) context.Context {
	return pkgauth.WithPrincipal(context.Background(),
		&domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff})
}

func line(sku string, system int32, counted *int32) store.CycleCountLine {
	l := store.CycleCountLine{ID: uuid.New(), Sku: sku, SystemQty: system}
	if counted != nil {
		l.CountedQty = pgtype.Int4{Int32: *counted, Valid: true}
	}
	return l
}

func intp(v int32) *int32 { return &v }

func TestCounts_CreateIsBlind(t *testing.T) {
	orgID, storeID, zoneID := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{lines: []store.CycleCountLine{line("SKU-1", 9, nil)}}
	e := echo.New()
	e.HideBanner = true
	h := counts.New(fs, nil)

	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/counts",
		strings.NewReader(`{"zone_id":"`+zoneID.String()+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var out counts.Session
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Lines, 1)
	assert.Equal(t, "SKU-1", out.Lines[0].SKU)
	// blind: system qty hidden while open
	assert.EqualValues(t, 0, out.Lines[0].SystemQty)
}

func TestCounts_CompleteReconcilesVariance(t *testing.T) {
	orgID, storeID, zoneID := uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{
		count: store.CycleCount{ID: uuid.New(), OrgID: orgID, StoreID: storeID, ZoneID: zoneID, Status: "open"},
		lines: []store.CycleCountLine{
			line("SKU-A", 10, intp(7)), // shrinkage -3 → reconcile
			line("SKU-B", 5, intp(5)),  // match → no change
			line("SKU-C", 4, nil),      // unentered → skip
		},
	}
	aud := &fakeAuditor{}
	e := echo.New()
	e.HideBanner = true
	h := counts.New(fs, aud)

	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/counts/"+fs.count.ID.String()+"/complete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), fs.count.ID.String())

	require.NoError(t, h.Complete(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var out counts.CompleteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "completed", out.Status)
	assert.Equal(t, 1, out.Reconciled)
	require.Len(t, out.Variances, 1)
	assert.Equal(t, "SKU-A", out.Variances[0].SKU)
	assert.Equal(t, -3, out.Variances[0].Variance)

	// only the varying SKU is corrected, to the counted value
	require.Len(t, fs.setQtyArg, 1)
	assert.Equal(t, "SKU-A", fs.setQtyArg[0].Sku)
	assert.EqualValues(t, 7, fs.setQtyArg[0].Qty)
	require.Len(t, aud.entries, 1)
	assert.Equal(t, "inventory.cycle_count", aud.entries[0].Action)
	assert.True(t, fs.completed)
}

func TestCounts_CompleteConflictWhenDone(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{count: store.CycleCount{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "completed"}}
	e := echo.New()
	e.HideBanner = true
	h := counts.New(fs, nil)

	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/counts/"+fs.count.ID.String()+"/complete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), fs.count.ID.String())

	err := h.Complete(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusConflict, he.Code)
}
