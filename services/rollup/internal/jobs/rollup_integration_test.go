package jobs_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/live-rack/pkg/chstore"
)

// TestMaterializedViews_Integration proves the 5-minute and hourly rollups
// populate from raw scans via their materialized views. Skipped unless
// CLICKHOUSE_URL is set.
//
//	CLICKHOUSE_URL=http://liverack:liverack@localhost:8123 CLICKHOUSE_DB=liverack \
//	go test ./services/rollup/... -run Integration -v
func TestMaterializedViews_Integration(t *testing.T) {
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

	org, zone := uuid.New(), uuid.New()
	// 3 scans in one 5-minute bucket / hour: 2 picks, 1 invalid place.
	rows := []map[string]any{
		scanRow(org, zone, "pick", 1),
		scanRow(org, zone, "pick", 1),
		scanRow(org, zone, "place", 0),
	}
	if err := ch.Insert(ctx, "scan_events_raw", rows); err != nil {
		t.Fatalf("insert raw: %v", err)
	}

	// SummingMergeTree merges async; sum at query time for a deterministic read.
	gotPerf := scalar(t, ch, fmt.Sprintf(
		"SELECT sum(scans), sum(picks), sum(invalid) FROM zone_perf_5m WHERE org_id='%s' AND zone_id='%s' FORMAT TSV",
		org, zone))
	if gotPerf != "3\t2\t1" {
		t.Errorf("zone_perf_5m = %q, want 3\\t2\\t1", gotPerf)
	}

	gotHeat := scalar(t, ch, fmt.Sprintf(
		"SELECT sum(scans) FROM heatmap_hourly WHERE org_id='%s' AND zone_id='%s' FORMAT TSV", org, zone))
	if gotHeat != "3" {
		t.Errorf("heatmap_hourly scans = %q, want 3", gotHeat)
	}
}

func scanRow(org, zone uuid.UUID, action string, valid int) map[string]any {
	return map[string]any{
		"org_id": org.String(), "ts": "2026-06-01 09:30:00.000",
		"store_id": uuid.Nil.String(), "zone_id": zone.String(),
		"scanner_id": "ZBR", "sku": "LR-INT", "action": action, "valid": valid, "reason": "",
	}
}

func scalar(t *testing.T, ch *chstore.Client, q string) string {
	t.Helper()
	body, err := ch.Query(context.Background(), q)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	return strings.TrimSpace(string(body))
}
