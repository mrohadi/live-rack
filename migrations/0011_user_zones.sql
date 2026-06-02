-- +goose Up
-- +goose StatementBegin

-- Zone-scoped access: a user with one or more rows here may only see those
-- zones; a user with no rows has org-wide zone access (admins, managers).
CREATE TABLE user_zones (
    org_id    UUID NOT NULL REFERENCES orgs(id)  ON DELETE CASCADE,
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zone_id   UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, zone_id)
);

CREATE INDEX idx_user_zones_user ON user_zones (user_id);
CREATE INDEX idx_user_zones_org  ON user_zones (org_id);

ALTER TABLE user_zones ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_zones_tenant ON user_zones
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE user_zones TO liverack_app;
GRANT ALL ON TABLE user_zones TO liverack_svc;

-- Tighten the zones policy: tenant match AND (user is unscoped OR zone assigned).
-- current_setting(..., true) yields NULL when unset (service tokens) -> unscoped.
DROP POLICY zones_tenant ON zones;

CREATE POLICY zones_tenant ON zones
    USING (
        org_id = current_setting('app.org_id')::uuid
        AND (
            NOT EXISTS (
                SELECT 1 FROM user_zones uz
                WHERE uz.user_id = current_setting('app.user_id', true)::uuid
            )
            OR id IN (
                SELECT uz.zone_id FROM user_zones uz
                WHERE uz.user_id = current_setting('app.user_id', true)::uuid
            )
        )
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS zones_tenant ON zones;
CREATE POLICY zones_tenant ON zones
    USING (org_id = current_setting('app.org_id')::uuid);
DROP TABLE IF EXISTS user_zones CASCADE;
-- +goose StatementEnd
