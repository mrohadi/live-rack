package chsink_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/ingest/internal/chsink"
)

type fakeInserter struct {
	table string
	rows  []map[string]any
	calls int
}

func (f *fakeInserter) Insert(_ context.Context, table string, rows []map[string]any) error {
	f.table = table
	f.rows = rows
	f.calls++
	return nil
}

func TestDecodeScan(t *testing.T) {
	org, zone, store := uuid.New(), uuid.New(), uuid.New()
	ts := time.Date(2026, 6, 1, 9, 30, 15, 0, time.UTC)
	data, _ := json.Marshal(events.ScanRecorded{
		OrgID: org, StoreID: store, ZoneID: zone, ScannerID: "ZBR-1",
		SKU: "LR-1240", Action: "pick", Valid: false, Reason: "wrong_zone", TS: ts,
	})

	row, err := chsink.DecodeScan(data)
	require.NoError(t, err)
	assert.Equal(t, org.String(), row["org_id"])
	assert.Equal(t, zone.String(), row["zone_id"])
	assert.Equal(t, "pick", row["action"])
	assert.Equal(t, 0, row["valid"])
	assert.Equal(t, "2026-06-01 09:30:15.000", row["ts"])
}

func TestDecodeScan_MissingOrg(t *testing.T) {
	data, _ := json.Marshal(events.ScanRecorded{SKU: "X"})
	_, err := chsink.DecodeScan(data)
	assert.Error(t, err)
}

func TestDecodeSale_DefaultsStoreAndTime(t *testing.T) {
	org := uuid.New()
	data, _ := json.Marshal(events.POSSale{OrgID: org, Source: "square", SKU: "SKU1", Qty: 3, AmountCents: 999})

	row, err := chsink.DecodeSale(data)
	require.NoError(t, err)
	assert.Equal(t, uuid.Nil.String(), row["store_id"])
	assert.Equal(t, int32(3), row["qty"])
	assert.NotEmpty(t, row["ts"])
}

func TestSink_HandleScan(t *testing.T) {
	org := uuid.New()
	data, _ := json.Marshal(events.ScanRecorded{OrgID: org, ZoneID: uuid.New(), SKU: "X", Action: "place", Valid: true})

	f := &fakeInserter{}
	require.NoError(t, chsink.New(f).HandleScan(context.Background(), data))
	assert.Equal(t, 1, f.calls)
	assert.Equal(t, "scan_events_raw", f.table)
	assert.Len(t, f.rows, 1)
	assert.Equal(t, 1, f.rows[0]["valid"])
}

func TestSink_HandleSale(t *testing.T) {
	org := uuid.New()
	data, _ := json.Marshal(events.POSSale{OrgID: org, Source: "shopify", SKU: "Y", Qty: 1, AmountCents: 500})

	f := &fakeInserter{}
	require.NoError(t, chsink.New(f).HandleSale(context.Background(), data))
	assert.Equal(t, "sales_events_raw", f.table)
	assert.Len(t, f.rows, 1)
}
