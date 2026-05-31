-- +goose Up
-- +goose StatementBegin

-- ─── Items master ─────────────────────────────────────────────────────────────
CREATE TABLE items (
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    sku        TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    category   TEXT        NOT NULL DEFAULT '',
    status     TEXT        NOT NULL DEFAULT 'active'
               CHECK (status IN ('active','discontinued','recalled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    UNIQUE (org_id, sku)
);
CREATE INDEX idx_items_org_category ON items (org_id, category);

-- ─── Item locations (current on-hand qty per zone) ────────────────────────────
CREATE TABLE item_locations (
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id   UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    zone_id    UUID        NOT NULL REFERENCES zones(id)  ON DELETE CASCADE,
    sku        TEXT        NOT NULL,
    qty        INTEGER     NOT NULL DEFAULT 0 CHECK (qty >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    UNIQUE (org_id, zone_id, sku)
);
CREATE INDEX idx_item_locations_store ON item_locations (org_id, store_id);
CREATE INDEX idx_item_locations_sku   ON item_locations (org_id, sku);

CREATE TRIGGER trg_items_updated_at
    BEFORE UPDATE ON items FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_item_locations_updated_at
    BEFORE UPDATE ON item_locations FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE items          ENABLE ROW LEVEL SECURITY;
ALTER TABLE item_locations ENABLE ROW LEVEL SECURITY;

CREATE POLICY items_tenant ON items
    USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY item_locations_tenant ON item_locations
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE items          TO liverack_app;
GRANT ALL ON TABLE items          TO liverack_svc;
GRANT ALL ON TABLE item_locations TO liverack_app;
GRANT ALL ON TABLE item_locations TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS item_locations CASCADE;
DROP TABLE IF EXISTS items CASCADE;
-- +goose StatementEnd
