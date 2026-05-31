-- +goose Up
-- +goose StatementBegin

-- A pipeline is an instantiated workflow scoped to a store (e.g. Item Restoration).
CREATE TABLE pipelines (
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id   UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    key        TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    UNIQUE (org_id, store_id, key)
);
CREATE INDEX idx_pipelines_store ON pipelines (org_id, store_id);

-- Ordered stage definitions for a pipeline. position is 0-based; sla_seconds = 0
-- means no SLA (terminal/parking stages).
CREATE TABLE pipeline_stages (
    id          UUID    NOT NULL DEFAULT gen_random_uuid(),
    org_id      UUID    NOT NULL REFERENCES orgs(id)      ON DELETE CASCADE,
    pipeline_id UUID    NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    position    INT     NOT NULL,
    name        TEXT    NOT NULL,
    sla_seconds BIGINT  NOT NULL DEFAULT 0 CHECK (sla_seconds >= 0),
    PRIMARY KEY (id),
    UNIQUE (pipeline_id, position)
);
CREATE INDEX idx_pipeline_stages_pipeline ON pipeline_stages (org_id, pipeline_id, position);

-- A card is a single item flowing through a pipeline. stage_position points at
-- its current stage; entered_stage_at resets on every move to drive ageing/SLA.
CREATE TABLE pipeline_cards (
    id               UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id           UUID        NOT NULL REFERENCES orgs(id)      ON DELETE CASCADE,
    pipeline_id      UUID        NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    stage_position   INT         NOT NULL DEFAULT 0,
    title            TEXT        NOT NULL,
    sku              TEXT        NOT NULL DEFAULT '',
    priority         TEXT        NOT NULL DEFAULT 'medium'
                     CHECK (priority IN ('low','medium','high')),
    owner_id         UUID                 REFERENCES users(id) ON DELETE SET NULL,
    entered_stage_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_pipeline_cards_pipeline ON pipeline_cards (org_id, pipeline_id, stage_position);

CREATE TRIGGER trg_pipelines_updated_at
    BEFORE UPDATE ON pipelines FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_pipeline_cards_updated_at
    BEFORE UPDATE ON pipeline_cards FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE pipelines       ENABLE ROW LEVEL SECURITY;
ALTER TABLE pipeline_stages ENABLE ROW LEVEL SECURITY;
ALTER TABLE pipeline_cards  ENABLE ROW LEVEL SECURITY;

CREATE POLICY pipelines_tenant ON pipelines
    USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY pipeline_stages_tenant ON pipeline_stages
    USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY pipeline_cards_tenant ON pipeline_cards
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE pipelines       TO liverack_app;
GRANT ALL ON TABLE pipelines       TO liverack_svc;
GRANT ALL ON TABLE pipeline_stages TO liverack_app;
GRANT ALL ON TABLE pipeline_stages TO liverack_svc;
GRANT ALL ON TABLE pipeline_cards  TO liverack_app;
GRANT ALL ON TABLE pipeline_cards  TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pipeline_cards  CASCADE;
DROP TABLE IF EXISTS pipeline_stages CASCADE;
DROP TABLE IF EXISTS pipelines       CASCADE;
-- +goose StatementEnd
