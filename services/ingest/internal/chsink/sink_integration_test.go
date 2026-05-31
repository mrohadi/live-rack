package chsink_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/live-rack/pkg/chstore"
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/ingest/internal/chsink"
)

// TestSink_Integration writes a scan and a sale into a live ClickHouse via the
// real chstore client and reads them back, validating column/type compat
// (DateTime64 text, UInt8 valid, UUID strings). Skipped unless CLICKHOUSE_URL set.
//
//	CLICKHOUSE_URL=http://liverack:liverack@localhost:8123 CLICKHOUSE_DB=liverack \
//	go test ./services/ingest/internal/chsink/ -run Integration -v
func TestSink_Integration(t *testing.T) {
	rawURL := os.Getenv("CLICKHOUSE_URL")
	if rawURL == "" {
		t.Skip("CLICKHOUSE_URL not set; skipping ClickHouse integration test")
	}
	db := os.Getenv("CLICKHOUSE_DB")
	if db == "" {
		db = "liverack"
	}
	cfg, err := chstore.ParseConfig(rawURL, db)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	ch := chstore.New(cfg)
	ctx := context.Background()
	if err := ch.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	org := uuid.New() // unique tenant isolates this run's rows
	sink := chsink.New(ch)

	scan, _ := json.Marshal(events.ScanRecorded{
		OrgID: org, StoreID: uuid.New(), ZoneID: uuid.New(), ScannerID: "ZBR-int",
		SKU: "LR-INT", Action: "pick", Valid: false, Reason: "wrong_zone", TS: time.Now(),
	})
	if err := sink.HandleScan(ctx, scan); err != nil {
		t.Fatalf("HandleScan: %v", err)
	}
	sale, _ := json.Marshal(events.POSSale{OrgID: org, Source: "square", SKU: "LR-INT", Qty: 2, AmountCents: 1000})
	if err := sink.HandleSale(ctx, sale); err != nil {
		t.Fatalf("HandleSale: %v", err)
	}

	assertCount(t, ch, "scan_events_raw", org)
	assertCount(t, ch, "sales_events_raw", org)
}

func assertCount(t *testing.T, ch *chstore.Client, table string, org uuid.UUID) {
	t.Helper()
	q := "SELECT count() FROM " + table + " WHERE org_id = '" + org.String() + "' FORMAT TSV"
	body, err := ch.Query(context.Background(), q)
	if err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got := strings.TrimSpace(string(body)); got != "1" {
		t.Errorf("%s count = %q, want 1", table, got)
	}
}
