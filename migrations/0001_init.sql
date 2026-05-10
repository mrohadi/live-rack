-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "timescaledb";

-- ─── Orgs ───────────────────────────────────────────────────────────────────
CREATE TABLE orgs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_org_id TEXT      NOT NULL UNIQUE,
    name       TEXT        NOT NULL,
    plan       TEXT        NOT NULL DEFAULT 'free' CHECK (plan IN ('free','starter','growth','enterprise')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Stores ─────────────────────────────────────────────────────────────────
CREATE TABLE stores (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    address    TEXT,
    lat        DOUBLE PRECISION,
    lon        DOUBLE PRECISION,
    timezone   TEXT        NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_stores_org_id ON stores(org_id);

-- ─── Users ──────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    clerk_user_id TEXT       NOT NULL UNIQUE,
    email        TEXT        NOT NULL,
    display_name TEXT        NOT NULL,
    avatar_url   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_org_id   ON users(org_id);
CREATE INDEX idx_users_clerk_id ON users(clerk_user_id);

-- ─── Roles ──────────────────────────────────────────────────────────────────
CREATE TABLE roles (
    id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name   TEXT NOT NULL CHECK (name IN ('admin','manager','staff','readonly','service')),
    UNIQUE (org_id, name)
);
CREATE INDEX idx_roles_org_id ON roles(org_id);

-- Seed built-in roles for every new org via trigger
CREATE OR REPLACE FUNCTION seed_org_roles() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO roles (org_id, name)
    VALUES
        (NEW.id, 'admin'),
        (NEW.id, 'manager'),
        (NEW.id, 'staff'),
        (NEW.id, 'readonly'),
        (NEW.id, 'service');
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_seed_org_roles
    AFTER INSERT ON orgs
    FOR EACH ROW EXECUTE FUNCTION seed_org_roles();

-- ─── Role bindings ──────────────────────────────────────────────────────────
CREATE TABLE role_bindings (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id  UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    UNIQUE (user_id, role_id)
);
CREATE INDEX idx_role_bindings_org_id  ON role_bindings(org_id);
CREATE INDEX idx_role_bindings_user_id ON role_bindings(user_id);

-- ─── Zone scopes (per-user store restriction) ────────────────────────────────
CREATE TABLE zone_scopes (
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id   UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    UNIQUE (user_id, store_id)
);
CREATE INDEX idx_zone_scopes_user_id ON zone_scopes(user_id);

-- ─── updated_at trigger ──────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$;

CREATE TRIGGER trg_orgs_updated_at   BEFORE UPDATE ON orgs   FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_stores_updated_at BEFORE UPDATE ON stores FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_users_updated_at  BEFORE UPDATE ON users  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─── Row Level Security ──────────────────────────────────────────────────────
ALTER TABLE orgs        ENABLE ROW LEVEL SECURITY;
ALTER TABLE stores      ENABLE ROW LEVEL SECURITY;
ALTER TABLE users       ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles       ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_bindings ENABLE ROW LEVEL SECURITY;
ALTER TABLE zone_scopes ENABLE ROW LEVEL SECURITY;

-- app_org_id is set per-request by the Go gateway before any query.
CREATE POLICY orgs_tenant        ON orgs        USING (id         = current_setting('app.org_id')::uuid);
CREATE POLICY stores_tenant      ON stores      USING (org_id     = current_setting('app.org_id')::uuid);
CREATE POLICY users_tenant       ON users       USING (org_id     = current_setting('app.org_id')::uuid);
CREATE POLICY roles_tenant       ON roles       USING (org_id     = current_setting('app.org_id')::uuid);
CREATE POLICY role_bindings_tenant ON role_bindings USING (org_id = current_setting('app.org_id')::uuid);
CREATE POLICY zone_scopes_tenant ON zone_scopes USING (org_id     = current_setting('app.org_id')::uuid);

-- Service role bypasses RLS (used by migrations, seed, rollup workers)
CREATE ROLE liverack_app  LOGIN PASSWORD 'changeme' NOINHERIT;
CREATE ROLE liverack_svc  LOGIN PASSWORD 'changeme' NOINHERIT;
GRANT ALL ON ALL TABLES    IN SCHEMA public TO liverack_app;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO liverack_app;
ALTER ROLE liverack_svc BYPASSRLS;
GRANT ALL ON ALL TABLES    IN SCHEMA public TO liverack_svc;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO liverack_svc;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS zone_scopes, role_bindings, roles, users, stores, orgs CASCADE;
DROP FUNCTION IF EXISTS seed_org_roles CASCADE;
DROP FUNCTION IF EXISTS set_updated_at CASCADE;
DROP ROLE IF EXISTS liverack_app, liverack_svc;
-- +goose StatementEnd
