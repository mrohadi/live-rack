-- name: GetOrgByIdpID :one
SELECT * FROM orgs WHERE idp_org_id = $1;

-- name: UpsertOrg :one
INSERT INTO orgs (idp_org_id, name, plan)
VALUES ($1, $2, $3)
ON CONFLICT (idp_org_id) DO UPDATE
    SET name = EXCLUDED.name, updated_at = NOW()
RETURNING *;

-- name: GetUserByIdpID :one
SELECT * FROM users WHERE idp_user_id = $1;

-- name: UpsertUser :one
INSERT INTO users (org_id, idp_user_id, email, display_name, avatar_url)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (idp_user_id) DO UPDATE
    SET email        = COALESCE(NULLIF(EXCLUDED.email, ''), users.email),
        display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), users.display_name),
        avatar_url   = COALESCE(NULLIF(EXCLUDED.avatar_url, ''), users.avatar_url),
        status       = CASE WHEN users.status = 'pending' THEN 'active' ELSE users.status END,
        updated_at   = NOW()
RETURNING *;

-- name: CreateInvitedUser :one
INSERT INTO users (org_id, idp_user_id, email, display_name, status)
VALUES ($1, $2, $3, $4, 'pending')
ON CONFLICT (idp_user_id) DO UPDATE
    SET email        = EXCLUDED.email,
        display_name = EXCLUDED.display_name,
        updated_at   = NOW()
RETURNING *;

-- name: GetUserRole :one
SELECT r.name
FROM role_bindings rb
JOIN roles r ON r.id = rb.role_id
WHERE rb.user_id = $1 AND rb.org_id = $2
LIMIT 1;

-- name: GetUserStoreIDs :many
SELECT store_id FROM zone_scopes WHERE user_id = $1 AND org_id = $2;

-- name: BindUserRole :exec
INSERT INTO role_bindings (org_id, user_id, role_id)
SELECT $1, $2, r.id FROM roles r WHERE r.org_id = $1 AND r.name = $3
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: ListStoresByOrg :many
SELECT * FROM stores WHERE org_id = $1 ORDER BY name;

-- name: CreateStore :one
INSERT INTO stores (org_id, name, address, lat, lon, timezone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetStore :one
SELECT * FROM stores WHERE id = $1 AND org_id = $2;
