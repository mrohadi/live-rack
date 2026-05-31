-- Rollup target tables + materialized views. The 5-minute zone-performance and
-- 7x24 heatmap rollups are maintained automatically by MVs on insert. The
-- time-to-sell and co-purchase-lift tables are written by the rollup service's
-- daily jobs (they need cross-row joins MVs cannot express).

CREATE TABLE IF NOT EXISTS zone_perf_5m (
    org_id  UUID,
    zone_id UUID,
    bucket  DateTime('UTC'),
    scans   UInt64,
    picks   UInt64,
    places  UInt64,
    invalid UInt64
) ENGINE = SummingMergeTree
ORDER BY (org_id, zone_id, bucket);

CREATE MATERIALIZED VIEW IF NOT EXISTS zone_perf_5m_mv TO zone_perf_5m AS
SELECT
    org_id,
    zone_id,
    toStartOfFiveMinutes(ts) AS bucket,
    count()                  AS scans,
    countIf(action = 'pick') AS picks,
    countIf(action = 'place') AS places,
    countIf(valid = 0)       AS invalid
FROM scan_events_raw
GROUP BY org_id, zone_id, bucket;

CREATE TABLE IF NOT EXISTS heatmap_hourly (
    org_id  UUID,
    zone_id UUID,
    dow     UInt8,
    hour    UInt8,
    scans   UInt64
) ENGINE = SummingMergeTree
ORDER BY (org_id, zone_id, dow, hour);

CREATE MATERIALIZED VIEW IF NOT EXISTS heatmap_hourly_mv TO heatmap_hourly AS
SELECT
    org_id,
    zone_id,
    toDayOfWeek(ts) AS dow,
    toHour(ts)      AS hour,
    count()         AS scans
FROM scan_events_raw
GROUP BY org_id, zone_id, dow, hour;

CREATE TABLE IF NOT EXISTS time_to_sell (
    org_id    UUID,
    day       Date,
    sku       String,
    avg_hours Float64,
    samples   UInt64
) ENGINE = ReplacingMergeTree
ORDER BY (org_id, day, sku);

CREATE TABLE IF NOT EXISTS sell_through (
    org_id     UUID,
    day        Date,
    sku        String,
    placed     UInt64,
    sold       UInt64,
    rate       Float64
) ENGINE = ReplacingMergeTree
ORDER BY (org_id, day, sku);

CREATE TABLE IF NOT EXISTS combos_lift (
    org_id      UUID,
    day         Date,
    sku_a       String,
    sku_b       String,
    pair_orders UInt64,
    lift        Float64
) ENGINE = ReplacingMergeTree
ORDER BY (org_id, day, sku_a, sku_b);
