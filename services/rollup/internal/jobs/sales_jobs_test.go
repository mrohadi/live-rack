package jobs_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/live-rack/pkg/chstore"
	"github.com/live-rack/services/rollup/internal/jobs"
)

func TestTimeToSell_SQL(t *testing.T) {
	sql := jobs.TimeToSell{}.SQL(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	for _, want := range []string{"INSERT INTO time_to_sell", "toDate('2026-06-01')", "first_place"} {
		if !strings.Contains(sql, want) {
			t.Errorf("SQL missing %q", want)
		}
	}
}

func TestSellThrough_SQL(t *testing.T) {
	sql := jobs.SellThrough{}.SQL(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	for _, want := range []string{"INSERT INTO sell_through", "sumIf(cnt, kind = 'placed')", "sold / placed"} {
		if !strings.Contains(sql, want) {
			t.Errorf("SQL missing %q", want)
		}
	}
}

// TestSalesJobs_Integration seeds a place scan + sale and verifies both daily
// jobs produce correct rows. Skipped unless CLICKHOUSE_URL is set.
func TestSalesJobs_Integration(t *testing.T) {
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

	org := uuid.New()
	zone := uuid.New()
	// Place at 08:00, two placed scans; sale of 1 unit at 12:00 (4h later).
	scans := []map[string]any{
		{"org_id": org.String(), "ts": "2026-06-01 08:00:00.000", "store_id": uuid.Nil.String(),
			"zone_id": zone.String(), "scanner_id": "Z", "sku": "LR-TTS", "action": "place", "valid": 1, "reason": ""},
		{"org_id": org.String(), "ts": "2026-06-01 08:05:00.000", "store_id": uuid.Nil.String(),
			"zone_id": zone.String(), "scanner_id": "Z", "sku": "LR-TTS", "action": "place", "valid": 1, "reason": ""},
	}
	if err := ch.Insert(ctx, "scan_events_raw", scans); err != nil {
		t.Fatalf("insert scans: %v", err)
	}
	sale := []map[string]any{
		{"org_id": org.String(), "ts": "2026-06-01 12:00:00.000", "store_id": uuid.Nil.String(),
			"source": "square", "order_id": "o1", "sku": "LR-TTS", "qty": 1, "amount_cents": 500, "currency": "USD", "channel": "pos"},
	}
	if err := ch.Insert(ctx, "sales_events_raw", sale); err != nil {
		t.Fatalf("insert sale: %v", err)
	}

	day := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	runner := jobs.NewRunner(ch, jobs.TimeToSell{}, jobs.SellThrough{})
	if err := runner.Run(ctx, day); err != nil {
		t.Fatalf("Run: %v", err)
	}

	tts := scalarQ(t, ch, fmt.Sprintf(
		"SELECT avg_hours, samples FROM time_to_sell FINAL WHERE org_id='%s' AND sku='LR-TTS' FORMAT TSV", org))
	if tts != "4\t1" {
		t.Errorf("time_to_sell = %q, want 4\\t1", tts)
	}

	st := scalarQ(t, ch, fmt.Sprintf(
		"SELECT placed, sold, rate FROM sell_through FINAL WHERE org_id='%s' AND sku='LR-TTS' FORMAT TSV", org))
	if st != "2\t1\t0.5" {
		t.Errorf("sell_through = %q, want 2\\t1\\t0.5", st)
	}
}

func scalarQ(t *testing.T, ch *chstore.Client, q string) string {
	t.Helper()
	body, err := ch.Query(context.Background(), q)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	return strings.TrimSpace(string(body))
}
