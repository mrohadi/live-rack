-- +goose Up
-- +goose StatementBegin

-- A cycle-count session snapshots on-hand qty per SKU in a zone, captures a
-- blind physical count, then reconciles variances back into item_locations.
CREATE TABLE cycle_counts (
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id     UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    zone_id      UUID        NOT NULL REFERENCES zones(id)  ON DELETE CASCADE,
    status       TEXT        NOT NULL DEFAULT 'open'
                 CHECK (status IN ('open','completed')),
    created_by   UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_cycle_counts_store ON cycle_counts (org_id, store_id, status);

-- One line per SKU snapshotted at session start. counted_qty is NULL until the
-- counter enters it (blind — system_qty is hidden in the count UI).
CREATE TABLE cycle_count_lines (
    id          UUID    NOT NULL DEFAULT gen_random_uuid(),
    count_id    UUID    NOT NULL REFERENCES cycle_counts(id) ON DELETE CASCADE,
    org_id      UUID    NOT NULL REFERENCES orgs(id)         ON DELETE CASCADE,
    sku         TEXT    NOT NULL,
    system_qty  INTEGER NOT NULL,
    counted_qty INTEGER,
    PRIMARY KEY (id),
    UNIQUE (count_id, sku)
);
CREATE INDEX idx_cycle_count_lines_count ON cycle_count_lines (count_id);

ALTER TABLE cycle_counts      ENABLE ROW LEVEL SECURITY;
ALTER TABLE cycle_count_lines ENABLE ROW LEVEL SECURITY;

CREATE POLICY cycle_counts_tenant ON cycle_counts
    USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY cycle_count_lines_tenant ON cycle_count_lines
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE cycle_counts      TO liverack_app;
GRANT ALL ON TABLE cycle_counts      TO liverack_svc;
GRANT ALL ON TABLE cycle_count_lines TO liverack_app;
GRANT ALL ON TABLE cycle_count_lines TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cycle_count_lines CASCADE;
DROP TABLE IF EXISTS cycle_counts CASCADE;
-- +goose StatementEnd
