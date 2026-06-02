package scans_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/scans"
)

// fakeZoneGetter implements scans.ZoneGetter.
type fakeZoneGetter struct {
	rows map[uuid.UUID]store.Zone
}

func (f *fakeZoneGetter) GetZone(_ context.Context, arg store.GetZoneParams) (store.Zone, error) {
	z, ok := f.rows[arg.ID]
	if !ok || z.OrgID != arg.OrgID {
		return store.Zone{}, errNotFound
	}
	return z, nil
}

type fakeRecorder struct {
	mu     sync.Mutex
	events []store.CreateScanEventParams
}

func (f *fakeRecorder) CreateScanEvent(_ context.Context, arg store.CreateScanEventParams) (store.ScanEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, arg)
	return store.ScanEvent{ID: uuid.New()}, nil
}

type fakeAdjuster struct {
	mu      sync.Mutex
	adjusts []store.AdjustItemLocationQtyParams
}

func (f *fakeAdjuster) AdjustItemLocationQty(_ context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.adjusts = append(f.adjusts, arg)
	return store.ItemLocation{ID: uuid.New(), Sku: arg.Sku, Qty: arg.Qty}, nil
}

// fakeRestocker implements scans.Restocker. items maps SKU→reorder_point.
type fakeRestocker struct {
	mu          sync.Mutex
	items       map[string]int32
	openByTitle map[string]int64
	created     []store.CreateTaskParams
}

func (f *fakeRestocker) GetItemBySKU(_ context.Context, arg store.GetItemBySKUParams) (store.Item, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rp, ok := f.items[arg.Sku]
	if !ok {
		return store.Item{}, errNotFound
	}
	return store.Item{Sku: arg.Sku, ReorderPoint: rp}, nil
}

func (f *fakeRestocker) CountOpenTasksByTitle(_ context.Context, arg store.CountOpenTasksByTitleParams) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.openByTitle[arg.Title], nil
}

func (f *fakeRestocker) CreateTask(_ context.Context, arg store.CreateTaskParams) (store.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, arg)
	return store.Task{ID: uuid.New()}, nil
}

type fakePublisher struct {
	mu       sync.Mutex
	subjects []string
	payloads []any
}

func (f *fakePublisher) Publish(_ context.Context, subject string, v any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.subjects = append(f.subjects, subject)
	f.payloads = append(f.payloads, v)
	return nil
}

var errNotFound = &notFoundError{}

type notFoundError struct{}

func (e *notFoundError) Error() string { return "zone not found" }

func withPrincipal(c echo.Context, orgID, storeID uuid.UUID) {
	p := &domain.Principal{
		UserID:   uuid.New(),
		OrgID:    orgID,
		IDPOrgID: "idp_" + orgID.String(),
		Role:     domain.RoleAdmin,
		StoreIDs: []uuid.UUID{storeID},
	}
	ctx := pkgauth.WithPrincipal(c.Request().Context(), p)
	c.SetRequest(c.Request().WithContext(ctx))
}

func mustConstraints(t *testing.T, c domain.ZoneConstraints) []byte {
	t.Helper()
	b, err := domain.MarshalConstraints(c)
	require.NoError(t, err)
	return b
}

// run wires one validate request and returns the recorder.
func run(t *testing.T, getter *fakeZoneGetter, orgID, storeID uuid.UUID, body map[string]any) (*httptest.ResponseRecorder, *fakeRecorder, *fakeAdjuster, *fakePublisher, error) {
	t.Helper()
	e := echo.New()
	e.HideBanner = true
	rec0 := &fakeRecorder{}
	adj := &fakeAdjuster{}
	pub := &fakePublisher{}
	// Empty items map → no reorder policy → auto-restock is a no-op here.
	restock := &fakeRestocker{items: map[string]int32{}, openByTitle: map[string]int64{}}
	h := scans.New(getter, rec0, adj, restock, pub)
	h.Register(e.Group("/api/v1/stores"))

	raw, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/scan/validate", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	withPrincipal(c, orgID, storeID)
	return rec, rec0, adj, pub, h.Validate(c)
}

func TestScanHandler_Validate(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	dwell := 60

	// Seed one zone: denies hazmat, capacity 10, dwell 60s, dual-scan on.
	zoneID := uuid.New()
	getter := &fakeZoneGetter{rows: map[uuid.UUID]store.Zone{
		zoneID: {
			ID:       zoneID,
			OrgID:    orgID,
			StoreID:  storeID,
			Capacity: 10,
			Constraints: mustConstraints(t, domain.ZoneConstraints{
				DeniedCategories: []string{"hazmat"},
				DwellSeconds:     &dwell,
				RequireDualScan:  true,
			}),
		},
	}}

	now := time.Now()

	cases := []struct {
		name             string
		body             map[string]any
		wantValid        bool
		wantCode         string
		wantRequiresDual bool
	}{
		{
			name:      "denied category",
			body:      map[string]any{"zone_id": zoneID, "category": "hazmat", "scan_qty": 1, "dual_scan_confirmed": true, "scanner_id": "scn-1", "sku": "SKU1", "action": "place"},
			wantValid: false,
			wantCode:  "category_denied",
		},
		{
			name:      "capacity exceeded",
			body:      map[string]any{"zone_id": zoneID, "category": "frozen", "current_qty": 10, "scan_qty": 1, "dual_scan_confirmed": true, "scanner_id": "scn-1", "sku": "SKU1", "action": "place"},
			wantValid: false,
			wantCode:  "capacity_exceeded",
		},
		{
			name:      "dwell violation",
			body:      map[string]any{"zone_id": zoneID, "category": "frozen", "scan_qty": 1, "last_scan_at": now.Add(-10 * time.Second), "dual_scan_confirmed": true, "scanner_id": "scn-1", "sku": "SKU1", "action": "place"},
			wantValid: false,
			wantCode:  "dwell_violation",
		},
		{
			name:             "dual-scan unconfirmed",
			body:             map[string]any{"zone_id": zoneID, "category": "frozen", "scan_qty": 1, "scanner_id": "scn-1", "sku": "SKU1", "action": "place"},
			wantValid:        false,
			wantCode:         "dual_scan_required",
			wantRequiresDual: true,
		},
		{
			name:      "valid scan",
			body:      map[string]any{"zone_id": zoneID, "category": "frozen", "scan_qty": 1, "dual_scan_confirmed": true, "scanner_id": "scn-1", "sku": "SKU1", "action": "place"},
			wantValid: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec, recorder, adj, pub, err := run(t, getter, orgID, storeID, tc.body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			// every scan persisted + published once
			assert.Len(t, recorder.events, 1)
			require.Len(t, pub.subjects, 1)
			assert.Equal(t, events.ScanSubject(orgID), pub.subjects[0])
			ev := pub.payloads[0].(events.ScanRecorded)
			assert.Equal(t, tc.wantValid, ev.Valid)
			// only valid scans adjust on-hand inventory
			if tc.wantValid {
				assert.Len(t, adj.adjusts, 1)
			} else {
				assert.Empty(t, adj.adjusts)
			}

			var resp struct {
				Valid            bool   `json:"valid"`
				Code             string `json:"code"`
				RequiresDualScan bool   `json:"requires_dual_scan"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tc.wantValid, resp.Valid)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantRequiresDual, resp.RequiresDualScan)
		})
	}
}

func TestScanHandler_Validate_ZoneNotFound(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	getter := &fakeZoneGetter{rows: map[uuid.UUID]store.Zone{}}

	_, _, _, _, err := run(t, getter, orgID, storeID, map[string]any{
		"zone_id": uuid.New(), "category": "frozen", "scan_qty": 1, "scanner_id": "scn-1", "sku": "SKU1", "action": "place",
	})
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func TestScanHandler_Validate_BadScanQty(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	getter := &fakeZoneGetter{rows: map[uuid.UUID]store.Zone{}}

	_, _, _, _, err := run(t, getter, orgID, storeID, map[string]any{
		"zone_id": uuid.New(), "category": "frozen", "scan_qty": 0, "scanner_id": "scn-1", "sku": "SKU1", "action": "place",
	})
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

// fixedAdjuster reports a constant resulting on-hand qty after each adjust.
type fixedAdjuster struct{ qty int32 }

func (f *fixedAdjuster) AdjustItemLocationQty(_ context.Context, arg store.AdjustItemLocationQtyParams) (store.ItemLocation, error) {
	return store.ItemLocation{ID: uuid.New(), Sku: arg.Sku, Qty: f.qty}, nil
}

// runRestock wires a validate request with custom adjuster + restocker.
func runRestock(
	t *testing.T,
	getter *fakeZoneGetter,
	adj scans.LocationAdjuster,
	restock *fakeRestocker,
	orgID, storeID uuid.UUID,
	body map[string]any,
) error {
	t.Helper()
	e := echo.New()
	e.HideBanner = true
	h := scans.New(getter, &fakeRecorder{}, adj, restock, &fakePublisher{})
	h.Register(e.Group("/api/v1/stores"))

	raw, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/scan/validate", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())
	withPrincipal(c, orgID, storeID)
	return h.Validate(c)
}

func TestScanHandler_AutoRestock(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	zoneID := uuid.New()
	getter := &fakeZoneGetter{rows: map[uuid.UUID]store.Zone{
		zoneID: {ID: zoneID, OrgID: orgID, StoreID: storeID, Capacity: 100},
	}}
	body := map[string]any{
		"zone_id": zoneID, "category": "frozen", "scan_qty": 1,
		"scanner_id": "scn-1", "sku": "SKU-9", "action": "pick",
	}

	t.Run("creates restock task when qty drops to reorder point", func(t *testing.T) {
		restock := &fakeRestocker{
			items:       map[string]int32{"SKU-9": 5},
			openByTitle: map[string]int64{},
		}
		require.NoError(t, runRestock(t, getter, &fixedAdjuster{qty: 4}, restock, orgID, storeID, body))
		require.Len(t, restock.created, 1)
		assert.Equal(t, domain.RestockTaskTitle("SKU-9"), restock.created[0].Title)
		assert.Equal(t, string(domain.TaskPriorityHigh), restock.created[0].Priority)
		assert.Equal(t, string(domain.TaskStatusTodo), restock.created[0].Status)
	})

	t.Run("no task when above reorder point", func(t *testing.T) {
		restock := &fakeRestocker{
			items:       map[string]int32{"SKU-9": 5},
			openByTitle: map[string]int64{},
		}
		require.NoError(t, runRestock(t, getter, &fixedAdjuster{qty: 20}, restock, orgID, storeID, body))
		assert.Empty(t, restock.created)
	})

	t.Run("dedupes against an existing open restock task", func(t *testing.T) {
		restock := &fakeRestocker{
			items:       map[string]int32{"SKU-9": 5},
			openByTitle: map[string]int64{domain.RestockTaskTitle("SKU-9"): 1},
		}
		require.NoError(t, runRestock(t, getter, &fixedAdjuster{qty: 1}, restock, orgID, storeID, body))
		assert.Empty(t, restock.created)
	})

	t.Run("no task when reorder point disabled", func(t *testing.T) {
		restock := &fakeRestocker{
			items:       map[string]int32{"SKU-9": 0},
			openByTitle: map[string]int64{},
		}
		require.NoError(t, runRestock(t, getter, &fixedAdjuster{qty: 0}, restock, orgID, storeID, body))
		assert.Empty(t, restock.created)
	})
}
