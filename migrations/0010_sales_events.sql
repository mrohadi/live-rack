-- +goose Up
-- +goose StatementBegin

CREATE TABLE sales_events (
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    ts           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    org_id       UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id     UUID                 REFERENCES stores(id) ON DELETE SET NULL,
    source       TEXT        NOT NULL,
    order_id     TEXT        NOT NULL DEFAULT '',
    sku          TEXT        NOT NULL DEFAULT '',
    qty          INT         NOT NULL DEFAULT 1,
    amount_cents BIGINT      NOT NULL DEFAULT 0,
    currency     TEXT        NOT NULL DEFAULT 'USD',
    channel      TEXT        NOT NULL DEFAULT '',
    PRIMARY KEY (id, ts)
);

SELECT create_hypertable('sales_events', 'ts');

CREATE INDEX idx_sales_events_org_ts  ON sales_events (org_id, ts DESC);
CREATE INDEX idx_sales_events_sku_ts  ON sales_events (org_id, sku, ts DESC);

ALTER TABLE sales_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY sales_events_tenant ON sales_events
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE sales_events TO liverack_app;
GRANT ALL ON TABLE sales_events TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sales_events CASCADE;
-- +goose StatementEnd
