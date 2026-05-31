-- Raw analytics tables. The ingest worker streams NATS events here; rollups and
-- materialized views aggregate downstream. Every table carries org_id first in
-- the sort key to keep tenant scans cheap and contiguous.

CREATE TABLE IF NOT EXISTS scan_events_raw (
    org_id     UUID,
    ts         DateTime64(3, 'UTC'),
    store_id   UUID,
    zone_id    UUID,
    scanner_id String,
    sku        String,
    action     LowCardinality(String),
    valid      UInt8,
    reason     String
) ENGINE = MergeTree
ORDER BY (org_id, ts)
PARTITION BY toYYYYMM(ts);

CREATE TABLE IF NOT EXISTS sales_events_raw (
    org_id       UUID,
    ts           DateTime64(3, 'UTC'),
    store_id     UUID,
    source       LowCardinality(String),
    order_id     String,
    sku          String,
    qty          Int32,
    amount_cents Int64,
    currency     LowCardinality(String),
    channel      LowCardinality(String)
) ENGINE = MergeTree
ORDER BY (org_id, ts)
PARTITION BY toYYYYMM(ts);
