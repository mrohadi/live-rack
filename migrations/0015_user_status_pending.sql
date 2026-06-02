-- +goose Up
-- +goose StatementBegin

-- Allow an invited-but-not-yet-joined lifecycle state so invitees appear in the
-- roster before their first login (which flips them to 'active' via UpsertUser).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE users
    ADD CONSTRAINT users_status_check
    CHECK (status IN ('active','break','off','pending'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE users
    ADD CONSTRAINT users_status_check
    CHECK (status IN ('active','break','off'));

-- +goose StatementEnd
