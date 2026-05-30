package scans_test

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

var errNotFound = &notFoundError{}

type notFoundError struct{}

func (e *notFoundError) Error() string { return "zone not found" }

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

func mustConstraints(t *testing.T, c domain.ZoneConstraints) []byte {
	t.Helper()
	b, err := domain.MarshalConstraints(c)
	require.NoError(t, err)
	return b
}

// run wires one validate request and returns the recorder.
func run(t *testing.T, getter *fakeZoneGetter, orgID, storeID uuid.UUID, body map[string]any) (*httptest.ResponseRecorder, error) {
	t.Helper()
	e := echo.New()
	e.HideBanner = true
	h := scans.New(getter)
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
	return rec, h.Validate(c)
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
			body:      map[string]any{"zone_id": zoneID, "category": "hazmat", "scan_qty": 1, "dual_scan_confirmed": true},
			wantValid: false,
			wantCode:  "category_denied",
		},
		{
			name:      "capacity exceeded",
			body:      map[string]any{"zone_id": zoneID, "category": "frozen", "current_qty": 10, "scan_qty": 1, "dual_scan_confirmed": true},
			wantValid: false,
			wantCode:  "capacity_exceeded",
		},
		{
			name:      "dwell violation",
			body:      map[string]any{"zone_id": zoneID, "category": "frozen", "scan_qty": 1, "last_scan_at": now.Add(-10 * time.Second), "dual_scan_confirmed": true},
			wantValid: false,
			wantCode:  "dwell_violation",
		},
		{
			name:             "dual-scan unconfirmed",
			body:             map[string]any{"zone_id": zoneID, "category": "frozen", "scan_qty": 1},
			wantValid:        false,
			wantCode:         "dual_scan_required",
			wantRequiresDual: true,
		},
		{
			name:      "valid scan",
			body:      map[string]any{"zone_id": zoneID, "category": "frozen", "scan_qty": 1, "dual_scan_confirmed": true},
			wantValid: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec, err := run(t, getter, orgID, storeID, tc.body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

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

	_, err := run(t, getter, orgID, storeID, map[string]any{
		"zone_id": uuid.New(), "category": "frozen", "scan_qty": 1,
	})
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func TestScanHandler_Validate_BadScanQty(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	getter := &fakeZoneGetter{rows: map[uuid.UUID]store.Zone{}}

	_, err := run(t, getter, orgID, storeID, map[string]any{
		"zone_id": uuid.New(), "category": "frozen", "scan_qty": 0,
	})
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}
