-- +goose Up
-- +goose StatementBegin

-- A pick list directs a picker through the store to fulfil an order. Lines are
-- sequenced (seq) into a map-optimised walking route at creation time; each line
-- is confirmed by scan, decrementing on-hand inventory.
CREATE TABLE pick_lists (
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id     UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    reference    TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL DEFAULT 'open'
                 CHECK (status IN ('open','picking','completed','cancelled')),
    created_by   UUID,
    assignee_id  UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_pick_lists_store ON pick_lists (org_id, store_id, status);

-- One line per SKU to pick. zone_id is the resolved source location; seq is the
-- optimised visit order (0-based). qty_picked + status are updated on confirm.
CREATE TABLE pick_list_lines (
    id            UUID    NOT NULL DEFAULT gen_random_uuid(),
    list_id       UUID    NOT NULL REFERENCES pick_lists(id) ON DELETE CASCADE,
    org_id        UUID    NOT NULL REFERENCES orgs(id)       ON DELETE CASCADE,
    zone_id       UUID    REFERENCES zones(id) ON DELETE SET NULL,
    sku           TEXT    NOT NULL,
    qty_requested INTEGER NOT NULL CHECK (qty_requested > 0),
    qty_picked    INTEGER NOT NULL DEFAULT 0 CHECK (qty_picked >= 0),
    seq           INTEGER NOT NULL DEFAULT 0,
    status        TEXT    NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending','picked','short')),
    PRIMARY KEY (id),
    UNIQUE (list_id, sku)
);
CREATE INDEX idx_pick_list_lines_list ON pick_list_lines (list_id, seq);

ALTER TABLE pick_lists      ENABLE ROW LEVEL SECURITY;
ALTER TABLE pick_list_lines ENABLE ROW LEVEL SECURITY;

CREATE POLICY pick_lists_tenant ON pick_lists
    USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY pick_list_lines_tenant ON pick_list_lines
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE pick_lists      TO liverack_app;
GRANT ALL ON TABLE pick_lists      TO liverack_svc;
GRANT ALL ON TABLE pick_list_lines TO liverack_app;
GRANT ALL ON TABLE pick_list_lines TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pick_list_lines CASCADE;
DROP TABLE IF EXISTS pick_lists CASCADE;
-- +goose StatementEnd
