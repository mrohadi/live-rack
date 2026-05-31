-- +goose Up
-- +goose StatementBegin

-- Append-only audit trail, range-partitioned by month so old periods can be
-- detached/archived cheaply. The writer creates the current month's partition
-- on demand; these initial partitions seed the table.
CREATE TABLE audit_log (
    id            UUID        NOT NULL DEFAULT gen_random_uuid(),
    ts            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    org_id        UUID        NOT NULL,
    actor_user_id UUID,
    action        TEXT        NOT NULL,
    resource_type TEXT        NOT NULL,
    resource_id   TEXT        NOT NULL DEFAULT '',
    metadata      JSONB       NOT NULL DEFAULT '{}',
    PRIMARY KEY (id, ts)
) PARTITION BY RANGE (ts);

CREATE INDEX idx_audit_log_org_ts ON audit_log (org_id, ts DESC);

CREATE TABLE audit_log_2026_06 PARTITION OF audit_log
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE audit_log_2026_07 PARTITION OF audit_log
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_log_tenant ON audit_log
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE audit_log TO liverack_app;
GRANT ALL ON TABLE audit_log TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_log CASCADE;
-- +goose StatementEnd
