-- +goose Up
-- +goose StatementBegin

CREATE TABLE scan_events (
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    ts         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    org_id     UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id   UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    zone_id    UUID        NOT NULL REFERENCES zones(id)  ON DELETE CASCADE,
    scanner_id TEXT        NOT NULL,
    sku        TEXT        NOT NULL,
    action     TEXT        NOT NULL DEFAULT 'place'
               CHECK (action IN ('place','pick','move','count')),
    valid      BOOLEAN     NOT NULL DEFAULT TRUE,
    reason     TEXT        NOT NULL DEFAULT '',
    PRIMARY KEY (id, ts)
);

SELECT create_hypertable('scan_events', 'ts');

CREATE INDEX idx_scan_events_zone_ts ON scan_events (org_id, zone_id, ts DESC);
CREATE INDEX idx_scan_events_sku_ts  ON scan_events (org_id, zone_id, sku, ts DESC);

ALTER TABLE scan_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY scan_events_tenant ON scan_events
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE scan_events TO liverack_app;
GRANT ALL ON TABLE scan_events TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS scan_events CASCADE;
-- +goose StatementEnd