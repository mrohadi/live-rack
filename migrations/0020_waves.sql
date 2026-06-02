-- +goose Up
-- +goose StatementBegin

-- A wave batches several pick lists into one merged walking route: the picker
-- collects the summed quantity per SKU+zone once, then the system allocates the
-- picked units back across the member orders (FIFO).
CREATE TABLE waves (
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id     UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    reference    TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL DEFAULT 'open'
                 CHECK (status IN ('open','picking','completed','cancelled')),
    created_by   UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_waves_store ON waves (org_id, store_id, status);

-- Membership: a pick list belongs to at most one wave.
ALTER TABLE pick_lists
    ADD COLUMN wave_id UUID REFERENCES waves(id) ON DELETE SET NULL;
CREATE INDEX idx_pick_lists_wave ON pick_lists (wave_id);

ALTER TABLE waves ENABLE ROW LEVEL SECURITY;
CREATE POLICY waves_tenant ON waves
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE waves TO liverack_app;
GRANT ALL ON TABLE waves TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pick_lists DROP COLUMN IF EXISTS wave_id;
DROP TABLE IF EXISTS waves CASCADE;
-- +goose StatementEnd
