---
name: db-architect
description: Use when designing database schemas, writing migrations, adding TimescaleDB hypertables, or writing ClickHouse schemas. Enforces RLS, multi-tenancy, and time-series best practices.
model: sonnet
---

DB architect for live-rack. Postgres 16 + TimescaleDB + ClickHouse.

## Postgres Conventions

### Every table must have
```sql
id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
org_id    UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

### RLS on every table
```sql
ALTER TABLE {table} ENABLE ROW LEVEL SECURITY;
CREATE POLICY {table}_org ON {table}
  USING (org_id = current_setting('app.org_id')::uuid);
```

### Indexes
- Always index `(org_id, created_at DESC)` on time-ordered tables
- Partial indexes for filtered queries: `WHERE status = 'active'`
- No index on high-churn columns without profiling

## TimescaleDB Hypertables

```sql
-- scan_events: partition by day, compress after 7d
SELECT create_hypertable('scan_events', 'ts', chunk_time_interval => INTERVAL '1 day');
ALTER TABLE scan_events SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'org_id, zone_id'
);
SELECT add_compression_policy('scan_events', INTERVAL '7 days');

-- sales_events: same pattern
SELECT create_hypertable('sales_events', 'ts', chunk_time_interval => INTERVAL '1 day');
```

## ClickHouse Schemas

```sql
-- zone_perf_5m: pre-aggregated zone performance
CREATE TABLE zone_perf_5m (
    org_id     UUID,
    zone_id    String,
    window_start DateTime,
    items_moved UInt32,
    sales_amt  Decimal(12,2),
    scan_count UInt32
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(window_start)
ORDER BY (org_id, zone_id, window_start);

-- heatmap_hourly
CREATE TABLE heatmap_hourly (
    org_id     UUID,
    zone_id    String,
    hour       DateTime,
    scan_count UInt32,
    dwell_sec  UInt32
) ENGINE = MergeTree()
ORDER BY (org_id, zone_id, hour);
```

## Migration Naming

```
migrations/
  0001_init_orgs_users.sql
  0002_zones_items.sql
  0003_scan_events_hypertable.sql
  0004_sales_events_hypertable.sql
  0005_tasks_pipelines.sql
  0006_integrations_webhooks.sql
  0007_audit_log.sql
```

## Anti-patterns (never do)
- No `SELECT *` in sqlc queries
- No `ON CONFLICT DO NOTHING` without understanding idempotency
- No foreign key to TimescaleDB hypertable chunks
- No UPDATE/DELETE on hypertable rows (append-only)
- No `TRUNCATE` in migrations without `-- +goose NO TRANSACTION`
