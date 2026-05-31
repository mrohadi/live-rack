-- +goose Up
-- +goose StatementBegin

CREATE TABLE integrations (
    id          UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    kind        TEXT        NOT NULL CHECK (kind IN ('shopify','square')),
    status      TEXT        NOT NULL DEFAULT 'disconnected'
                CHECK (status IN ('connected','disconnected','error')),
    external_id TEXT        NOT NULL DEFAULT '',
    secret      TEXT        NOT NULL DEFAULT '',
    config      JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    UNIQUE (org_id, kind)
);
-- Route inbound webhooks (which carry no org context) back to an org by vendor handle.
CREATE UNIQUE INDEX idx_integrations_kind_external ON integrations (kind, external_id)
    WHERE external_id <> '';

CREATE TRIGGER trg_integrations_updated_at
    BEFORE UPDATE ON integrations FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Idempotency + audit log of every inbound webhook delivery.
CREATE TABLE webhooks_inbound (
    id          UUID        NOT NULL DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    provider    TEXT        NOT NULL,
    event_id    TEXT        NOT NULL,
    topic       TEXT        NOT NULL DEFAULT '',
    status      TEXT        NOT NULL DEFAULT 'received'
                CHECK (status IN ('received','processed','rejected')),
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    UNIQUE (org_id, provider, event_id)
);
CREATE INDEX idx_webhooks_inbound_org_ts ON webhooks_inbound (org_id, received_at DESC);

ALTER TABLE integrations     ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhooks_inbound ENABLE ROW LEVEL SECURITY;

CREATE POLICY integrations_tenant ON integrations
    USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY webhooks_inbound_tenant ON webhooks_inbound
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE integrations     TO liverack_app;
GRANT ALL ON TABLE integrations     TO liverack_svc;
GRANT ALL ON TABLE webhooks_inbound TO liverack_app;
GRANT ALL ON TABLE webhooks_inbound TO liverack_svc;

-- Inbound webhooks carry no org context, so RLS would hide every integration row.
-- This SECURITY DEFINER resolver runs as the table owner to map a vendor handle
-- to its org + signing secret; the gateway then SETs app.org_id and proceeds.
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION resolve_webhook_integration(p_kind TEXT, p_external TEXT)
RETURNS TABLE (org_id UUID, secret TEXT)
LANGUAGE sql SECURITY DEFINER AS $$
    SELECT integrations.org_id, integrations.secret FROM public.integrations
    WHERE integrations.kind = p_kind AND integrations.external_id = p_external;
$$;
-- +goose StatementEnd
-- +goose StatementBegin
GRANT EXECUTE ON FUNCTION resolve_webhook_integration(TEXT, TEXT) TO liverack_app;
GRANT EXECUTE ON FUNCTION resolve_webhook_integration(TEXT, TEXT) TO liverack_svc;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS webhooks_inbound CASCADE;
DROP TABLE IF EXISTS integrations     CASCADE;
-- +goose StatementEnd
