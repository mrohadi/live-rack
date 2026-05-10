-- +goose Up
-- +goose StatementBegin

CREATE TYPE zone_type AS ENUM (
    'general', 'frozen', 'returns', 'staging', 'display', 'checkout'
);

CREATE TABLE zones (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id)    ON DELETE CASCADE,
    store_id    UUID        NOT NULL REFERENCES stores(id)  ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    type        zone_type   NOT NULL DEFAULT 'general',
    x           DOUBLE PRECISION NOT NULL DEFAULT 0,
    y           DOUBLE PRECISION NOT NULL DEFAULT 0,
    width       DOUBLE PRECISION NOT NULL DEFAULT 100,
    height      DOUBLE PRECISION NOT NULL DEFAULT 100,
    color       TEXT        NOT NULL DEFAULT '#6366f1',
    capacity    INT         NOT NULL DEFAULT 0 CHECK (capacity >= 0),
    constraints JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_zones_org_id   ON zones(org_id);
CREATE INDEX idx_zones_store_id ON zones(store_id);

CREATE TRIGGER trg_zones_updated_at
    BEFORE UPDATE ON zones
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE zones ENABLE ROW LEVEL SECURITY;

CREATE POLICY zones_tenant ON zones
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE zones TO liverack_app;
GRANT ALL ON TABLE zones TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS zones CASCADE;
DROP TYPE  IF EXISTS zone_type CASCADE;
-- +goose StatementEnd
