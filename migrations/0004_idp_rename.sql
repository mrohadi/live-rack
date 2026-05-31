-- +goose Up
-- Rename IdP-specific columns from Clerk to neutral idp_* (swap to Zitadel, LR-005a).
ALTER TABLE orgs  RENAME COLUMN clerk_org_id  TO idp_org_id;
ALTER TABLE users RENAME COLUMN clerk_user_id TO idp_user_id;

ALTER INDEX idx_users_clerk_id RENAME TO idx_users_idp_id;

-- +goose Down
ALTER INDEX idx_users_idp_id RENAME TO idx_users_clerk_id;

ALTER TABLE users RENAME COLUMN idp_user_id TO clerk_user_id;
ALTER TABLE orgs  RENAME COLUMN idp_org_id  TO clerk_org_id;
