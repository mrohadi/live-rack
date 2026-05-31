-- Demographic snapshots per store/zone. Used by the analytics layer to overlay
-- catchment-area signals (income, foot traffic, age mix) onto zone performance.
-- ReplacingMergeTree so re-loading a day's snapshot overwrites in place.

CREATE TABLE IF NOT EXISTS demographics (
    org_id   UUID,
    store_id UUID,
    zone_id  UUID,
    segment  LowCardinality(String),
    metric   LowCardinality(String),
    value    Float64,
    day      Date
) ENGINE = ReplacingMergeTree
ORDER BY (org_id, store_id, zone_id, segment, metric, day);
