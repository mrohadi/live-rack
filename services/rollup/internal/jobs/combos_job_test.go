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

func TestCombosLift_SQL(t *testing.T) {
	sql := jobs.CombosLift{}.SQL(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	for _, want := range []string{"INSERT INTO combos_lift", "a.sku < b.sku", "pair.pair_orders * t.n"} {
		if !strings.Contains(sql, want) {
			t.Errorf("SQL missing %q", want)
		}
	}
}

// TestCombosLift_Integration seeds three orders and asserts the lift for the
// (A,B) pair. Skipped unless CLICKHOUSE_URL is set.
func TestCombosLift_Integration(t *testing.T) {
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
	// orders: o1{A,B}, o2{A,B}, o3{C,D}. N=3, support A=B=2, C=D=1.
	// lift(A,B) = 2*3/(2*2) = 1.5
	var rows []map[string]any
	add := func(order, sku string) {
		rows = append(rows, map[string]any{
			"org_id": org.String(), "ts": "2026-06-01 10:00:00.000", "store_id": uuid.Nil.String(),
			"source": "square", "order_id": order, "sku": sku, "qty": 1, "amount_cents": 100, "currency": "USD", "channel": "pos",
		})
	}
	add("o1", "A")
	add("o1", "B")
	add("o2", "A")
	add("o2", "B")
	add("o3", "C")
	add("o3", "D")
	if err := ch.Insert(ctx, "sales_events_raw", rows); err != nil {
		t.Fatalf("insert: %v", err)
	}

	day := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if err := jobs.NewRunner(ch, jobs.CombosLift{}).Run(ctx, day); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := scalarQ(t, ch, fmt.Sprintf(
		"SELECT pair_orders, lift FROM combos_lift FINAL WHERE org_id='%s' AND sku_a='A' AND sku_b='B' FORMAT TSV", org))
	if got != "2\t1.5" {
		t.Errorf("combos_lift(A,B) = %q, want 2\\t1.5", got)
	}
}
