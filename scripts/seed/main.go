// seed loads dev fixtures (orgs, stores, zones, items, item_locations) into
// the local Postgres named by DATABASE_URL. Idempotent: re-running upserts
// the same rows. Run with: make seed
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

// fixtures is the full dev dataset as one idempotent SQL batch. Fixed UUIDs
// keep every run stable; ON CONFLICT clauses make re-seeding safe.
const fixtures = `
-- ─── Org + store ──────────────────────────────────────────────────────────
INSERT INTO orgs (id, idp_org_id, name, plan) VALUES
  ('11111111-1111-1111-1111-111111111111', 'dev-org', 'Acme Retail', 'growth')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, plan = EXCLUDED.plan;

INSERT INTO stores (id, org_id, name, address, timezone) VALUES
  ('22222222-2222-2222-2222-222222222222',
   '11111111-1111-1111-1111-111111111111',
   'Store #14', '500 Market St', 'America/New_York')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;

-- ─── Zones ────────────────────────────────────────────────────────────────
INSERT INTO zones (id, org_id, store_id, name, type, x, y, width, height, color, capacity) VALUES
  ('aaaa0001-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'Electronics Bay', 'general',  40,  40, 200, 140, '#6366f1', 500),
  ('aaaa0002-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'Frozen Aisle',   'frozen',  260,  40, 160, 140, '#0ea5e9', 300),
  ('aaaa0003-0000-0000-0000-000000000003', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'Returns Desk',   'returns',  40, 200, 160, 120, '#f59e0b', 100),
  ('aaaa0004-0000-0000-0000-000000000004', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'Checkout Front', 'checkout',220, 200, 200, 120, '#22c55e', 50)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, type = EXCLUDED.type;

-- ─── Items ────────────────────────────────────────────────────────────────
INSERT INTO items (org_id, sku, name, category, status) VALUES
  ('11111111-1111-1111-1111-111111111111', 'WID-001', 'Widget Pro',          'electronics', 'active'),
  ('11111111-1111-1111-1111-111111111111', 'WID-002', 'Wide Angle Lens',     'electronics', 'active'),
  ('11111111-1111-1111-1111-111111111111', 'GAD-010', 'Gadget Mini',         'electronics', 'active'),
  ('11111111-1111-1111-1111-111111111111', 'GAD-011', 'Gadget Max',          'electronics', 'active'),
  ('11111111-1111-1111-1111-111111111111', 'CAB-100', 'USB-C Cable 2m',      'accessories', 'active'),
  ('11111111-1111-1111-1111-111111111111', 'CAB-101', 'HDMI Cable 3m',       'accessories', 'active'),
  ('11111111-1111-1111-1111-111111111111', 'FRZ-200', 'Frozen Berries 1kg',  'grocery',     'active'),
  ('11111111-1111-1111-1111-111111111111', 'FRZ-201', 'Ice Cream Vanilla',   'grocery',     'active'),
  ('11111111-1111-1111-1111-111111111111', 'BAT-300', 'AA Batteries 8pk',    'accessories', 'active'),
  ('11111111-1111-1111-1111-111111111111', 'OLD-900', 'Legacy Remote',       'electronics', 'discontinued'),
  ('11111111-1111-1111-1111-111111111111', 'REC-901', 'Recalled Charger',    'accessories', 'recalled')
ON CONFLICT (org_id, sku) DO UPDATE SET name = EXCLUDED.name, category = EXCLUDED.category, status = EXCLUDED.status;

-- ─── Item locations (on-hand qty per zone) ──────────────────────────────────
INSERT INTO item_locations (org_id, store_id, zone_id, sku, qty) VALUES
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0001-0000-0000-0000-000000000001', 'WID-001', 42),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0001-0000-0000-0000-000000000001', 'WID-002', 18),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0001-0000-0000-0000-000000000001', 'GAD-010', 7),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0001-0000-0000-0000-000000000001', 'GAD-011', 3),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0001-0000-0000-0000-000000000001', 'CAB-100', 120),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0001-0000-0000-0000-000000000001', 'BAT-300', 64),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0002-0000-0000-0000-000000000002', 'FRZ-200', 30),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0002-0000-0000-0000-000000000002', 'FRZ-201', 25),
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'aaaa0003-0000-0000-0000-000000000003', 'REC-901', 1)
ON CONFLICT (org_id, zone_id, sku) DO UPDATE SET qty = EXCLUDED.qty;
`

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL not set")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		slog.Error("connect", "err", err)
		os.Exit(1)
	}

	defer func() {
		if err := conn.Close(ctx); err != nil {
			slog.Error("close db connection", "err", err)
		}
	}()

	if _, err := conn.Exec(ctx, fixtures); err != nil {
		slog.Error("seed", "err", err)
		os.Exit(1)
	}

	slog.Info("seed complete", "org", "Acme Retail", "store", "Store #14")
}
