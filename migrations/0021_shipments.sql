-- +goose Up
-- +goose StatementBegin

-- A shipment packs the picked goods from a pick list into a parcel, records the
-- carrier + tracking number, and is dispatched out of the building.
CREATE TABLE shipments (
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id          UUID        NOT NULL REFERENCES orgs(id)       ON DELETE CASCADE,
    store_id        UUID        NOT NULL REFERENCES stores(id)     ON DELETE CASCADE,
    pick_list_id    UUID        REFERENCES pick_lists(id) ON DELETE SET NULL,
    reference       TEXT        NOT NULL DEFAULT '',
    carrier         TEXT        NOT NULL DEFAULT '',
    tracking_number TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL DEFAULT 'packing'
                    CHECK (status IN ('packing','packed','dispatched','cancelled')),
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dispatched_at   TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_shipments_store ON shipments (org_id, store_id, status);

-- Line snapshot taken when the shipment is created from a pick list.
CREATE TABLE shipment_items (
    id          UUID    NOT NULL DEFAULT gen_random_uuid(),
    shipment_id UUID    NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    org_id      UUID    NOT NULL REFERENCES orgs(id)      ON DELETE CASCADE,
    sku         TEXT    NOT NULL,
    qty         INTEGER NOT NULL CHECK (qty > 0),
    PRIMARY KEY (id),
    UNIQUE (shipment_id, sku)
);
CREATE INDEX idx_shipment_items_shipment ON shipment_items (shipment_id);

ALTER TABLE shipments      ENABLE ROW LEVEL SECURITY;
ALTER TABLE shipment_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY shipments_tenant ON shipments
    USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY shipment_items_tenant ON shipment_items
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE shipments      TO liverack_app;
GRANT ALL ON TABLE shipments      TO liverack_svc;
GRANT ALL ON TABLE shipment_items TO liverack_app;
GRANT ALL ON TABLE shipment_items TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shipment_items CASCADE;
DROP TABLE IF EXISTS shipments CASCADE;
-- +goose StatementEnd
