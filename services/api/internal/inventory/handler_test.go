package inventory_test

import (
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
	"github.com/live-rack/services/api/internal/inventory"
)

type fakeLister struct {
	gotOrg   uuid.UUID
	gotStore uuid.UUID
	rows     []store.ListInventoryByStoreRow
}

func (f *fakeLister) ListInventoryByStore(_ context.Context, arg store.ListInventoryByStoreParams) ([]store.ListInventoryByStoreRow, error) {
	f.gotOrg = arg.OrgID
	f.gotStore = arg.StoreID
	return f.rows, nil
}

func TestInventoryHandler_List(t *testing.T) {
	orgID := uuid.New()
	storeID := uuid.New()
	zoneID := uuid.New()

	lister := &fakeLister{rows: []store.ListInventoryByStoreRow{{
		ID: uuid.New(), OrgID: orgID, StoreID: storeID, ZoneID: zoneID,
		Sku: "SKU-1", Qty: 7, Name: "Widget", Category: "frozen", Status: "active",
		Picks7d: 8, Picks30d: 20,
		UpdatedAt: time.Now().UTC(),
	}}}

	e := echo.New()
	e.HideBanner = true
	h := inventory.New(lister)
	h.Register(e.Group("/api/v1/stores"))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/stores/"+storeID.String()+"/inventory", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleAdmin}
	c.SetRequest(c.Request().WithContext(pkgauth.WithPrincipal(c.Request().Context(), p)))

	require.NoError(t, h.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, orgID, lister.gotOrg)
	assert.Equal(t, storeID, lister.gotStore)

	var out []inventory.Row
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "SKU-1", out[0].SKU)
	assert.EqualValues(t, 7, out[0].Qty)
	assert.Equal(t, "Widget", out[0].Name)
	assert.Equal(t, "hot", out[0].Velocity)
}
