-- +goose Up
-- +goose StatementBegin

-- Profile + presence fields backing the Users & Access screen.
ALTER TABLE users
    ADD COLUMN title        TEXT        NOT NULL DEFAULT '',
    ADD COLUMN shift        TEXT        NOT NULL DEFAULT 'day'
        CHECK (shift IN ('day','night','open','on-call')),
    ADD COLUMN status       TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','break','off')),
    ADD COLUMN mfa_enabled  BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN last_seen_at TIMESTAMPTZ;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS shift,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS mfa_enabled,
    DROP COLUMN IF EXISTS last_seen_at;

-- +goose StatementEnd
