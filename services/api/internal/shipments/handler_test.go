package shipments_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/shipments"
)

type fakeStore struct {
	list       store.PickList
	lines      []store.ListPickListLinesRow
	ship       store.Shipment
	items      []store.ListShipmentItemsRow
	added      []store.AddShipmentItemParams
	dispatch   store.MarkShipmentDispatchedParams
	dispatched bool
}

func (f *fakeStore) GetPickList(_ context.Context, _ store.GetPickListParams) (store.PickList, error) {
	if f.list.ID == uuid.Nil {
		return store.PickList{}, pgx.ErrNoRows
	}
	return f.list, nil
}
func (f *fakeStore) ListPickListLines(_ context.Context, _ uuid.UUID) ([]store.ListPickListLinesRow, error) {
	return f.lines, nil
}
func (f *fakeStore) CreateShipment(_ context.Context, arg store.CreateShipmentParams) (store.Shipment, error) {
	f.ship = store.Shipment{ID: uuid.New(), OrgID: arg.OrgID, StoreID: arg.StoreID, Reference: arg.Reference, Status: "packing"}
	return f.ship, nil
}
func (f *fakeStore) AddShipmentItem(_ context.Context, arg store.AddShipmentItemParams) (store.ShipmentItem, error) {
	f.added = append(f.added, arg)
	f.items = append(f.items, store.ListShipmentItemsRow{ID: uuid.New(), Sku: arg.Sku, Qty: arg.Qty})
	return store.ShipmentItem{ID: uuid.New(), Sku: arg.Sku, Qty: arg.Qty}, nil
}
func (f *fakeStore) GetShipment(_ context.Context, _ store.GetShipmentParams) (store.Shipment, error) {
	if f.ship.ID == uuid.Nil {
		return store.Shipment{}, pgx.ErrNoRows
	}
	return f.ship, nil
}
func (f *fakeStore) ListShipmentsByStore(_ context.Context, _ store.ListShipmentsByStoreParams) ([]store.ListShipmentsByStoreRow, error) {
	return nil, nil
}
func (f *fakeStore) ListShipmentItems(_ context.Context, _ uuid.UUID) ([]store.ListShipmentItemsRow, error) {
	return f.items, nil
}
func (f *fakeStore) MarkShipmentPacked(_ context.Context, _ store.MarkShipmentPackedParams) (store.Shipment, error) {
	f.ship.Status = "packed"
	return f.ship, nil
}
func (f *fakeStore) MarkShipmentDispatched(_ context.Context, arg store.MarkShipmentDispatchedParams) (store.Shipment, error) {
	f.dispatch = arg
	f.dispatched = true
	f.ship.Status = "dispatched"
	f.ship.Carrier = arg.Carrier
	f.ship.TrackingNumber = arg.TrackingNumber
	return f.ship, nil
}
func (f *fakeStore) CancelShipment(_ context.Context, _ store.CancelShipmentParams) (store.Shipment, error) {
	f.ship.Status = "cancelled"
	return f.ship, nil
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

func line(sku string, picked int32) store.ListPickListLinesRow {
	return store.ListPickListLinesRow{ID: uuid.New(), Sku: sku, QtyRequested: picked, QtyPicked: picked}
}

func TestShipments_CreateSnapshotsPickedLines(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{
		list:  store.PickList{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "completed"},
		lines: []store.ListPickListLinesRow{line("SKU-A", 3), line("SKU-B", 0), line("SKU-C", 2)},
	}
	aud := &fakeAuditor{}
	e := echo.New()
	e.HideBanner = true
	h := shipments.New(fs, aud)

	body := `{"pick_list_id":"` + fs.list.ID.String() + `","reference":"SHIP-1"}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/shipments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	// only picked lines (qty > 0) snapshot
	require.Len(t, fs.added, 2)
	assert.Equal(t, "SKU-A", fs.added[0].Sku)
	assert.Equal(t, "SKU-C", fs.added[1].Sku)

	var out shipments.Board
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Items, 2)
	require.Len(t, aud.entries, 1)
	assert.Equal(t, "inventory.shipment_create", aud.entries[0].Action)
}

func TestShipments_CreateRejectsIncompletePickList(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{list: store.PickList{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "picking"}}
	e := echo.New()
	e.HideBanner = true
	h := shipments.New(fs, nil)

	body := `{"pick_list_id":"` + fs.list.ID.String() + `"}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/shipments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusConflict, he.Code)
}

func TestShipments_DispatchRequiresPacked(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{ship: store.Shipment{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "packing"}}
	e := echo.New()
	e.HideBanner = true
	h := shipments.New(fs, nil)

	body := `{"carrier":"ups","tracking_number":"1Z"}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/shipments/"+fs.ship.ID.String()+"/dispatch",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), fs.ship.ID.String())

	err := h.Dispatch(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusConflict, he.Code)
}

func TestShipments_DispatchRecordsCarrier(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{ship: store.Shipment{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Status: "packed"}}
	aud := &fakeAuditor{}
	e := echo.New()
	e.HideBanner = true
	h := shipments.New(fs, aud)

	body := `{"carrier":"ups","tracking_number":"1Z999"}`
	req := httptest.NewRequestWithContext(ctxFor(orgID), http.MethodPost,
		"/api/v1/stores/"+storeID.String()+"/shipments/"+fs.ship.ID.String()+"/dispatch",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), fs.ship.ID.String())

	require.NoError(t, h.Dispatch(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, fs.dispatched)
	assert.Equal(t, "ups", fs.dispatch.Carrier)
	assert.Equal(t, "1Z999", fs.dispatch.TrackingNumber)

	var out shipments.Board
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "dispatched", out.Status)
	require.Len(t, aud.entries, 1)
	assert.Equal(t, "inventory.shipment_dispatch", aud.entries[0].Action)
}
