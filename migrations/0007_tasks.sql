-- +goose Up
-- +goose StatementBegin

CREATE TABLE tasks (
    id          UUID         NOT NULL DEFAULT gen_random_uuid(),
    org_id      UUID         NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    store_id    UUID         NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    zone_id     UUID                  REFERENCES zones(id)  ON DELETE SET NULL,
    title       TEXT         NOT NULL,
    status      TEXT         NOT NULL DEFAULT 'todo'
                CHECK (status IN ('todo','in_progress','review','done')),
    priority    TEXT         NOT NULL DEFAULT 'med'
                CHECK (priority IN ('low','med','high')),
    assignee_id UUID                  REFERENCES users(id)  ON DELETE SET NULL,
    due_at      TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_tasks_store_status ON tasks (org_id, store_id, status);
CREATE INDEX idx_tasks_assignee     ON tasks (org_id, assignee_id);

CREATE TRIGGER trg_tasks_updated_at
    BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
CREATE POLICY tasks_tenant ON tasks
    USING (org_id = current_setting('app.org_id')::uuid);

GRANT ALL ON TABLE tasks TO liverack_app;
GRANT ALL ON TABLE tasks TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks CASCADE;
-- +goose StatementEnd