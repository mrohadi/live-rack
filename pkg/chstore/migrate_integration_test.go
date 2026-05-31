package chstore_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/live-rack/pkg/chstore"
)

// TestMigrate_Integration applies the embedded schema to a live ClickHouse and
// asserts the expected tables exist. Skipped unless CLICKHOUSE_URL is set, so
// `go test ./...` stays green without a running warehouse.
//
//	CLICKHOUSE_URL=http://liverack:liverack@localhost:8123 \
//	CLICKHOUSE_DB=liverack go test ./pkg/chstore/ -run Integration -v
func TestMigrate_Integration(t *testing.T) {
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
	c := chstore.New(cfg)
	ctx := context.Background()

	if err := c.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if err := c.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	body, err := c.Query(ctx, "SHOW TABLES FORMAT TSV")
	if err != nil {
		t.Fatalf("SHOW TABLES: %v", err)
	}
	got := string(body)
	for _, want := range []string{
		"scan_events_raw", "sales_events_raw",
		"zone_perf_5m", "zone_perf_5m_mv",
		"heatmap_hourly", "heatmap_hourly_mv",
		"time_to_sell", "sell_through", "combos_lift",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing table %q in:\n%s", want, got)
		}
	}
}
