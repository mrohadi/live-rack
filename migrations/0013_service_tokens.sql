-- +goose Up
-- +goose StatementBegin

-- Service tokens are first-class principals (service role) authenticating via an
-- opaque bearer ("lrk_..."). Only the SHA-256 hash is stored.
CREATE TABLE service_tokens (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    token_hash   TEXT        NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX idx_service_tokens_org ON service_tokens (org_id);

ALTER TABLE service_tokens ENABLE ROW LEVEL SECURITY;

CREATE POLICY service_tokens_tenant ON service_tokens
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE service_tokens TO liverack_app;
GRANT ALL ON TABLE service_tokens TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS service_tokens CASCADE;
-- +goose StatementEnd
