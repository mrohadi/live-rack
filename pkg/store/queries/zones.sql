-- name: CreateZone :one
INSERT INTO zones (org_id, store_id, name, type, x, y, width, height, color, capacity, constraints)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetZone :one
SELECT * FROM zones WHERE id = $1 AND org_id = $2;

-- name: ListZonesByStore :many
SELECT * FROM zones WHERE store_id = $1 AND org_id = $2 ORDER BY name;

-- name: UpdateZone :one
UPDATE zones
SET name        = $3,
    type        = $4,
    x           = $5,
    y           = $6,
    width       = $7,
    height      = $8,
    color       = $9,
    capacity    = $10,
    constraints = $11
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteZone :exec
DELETE FROM zones WHERE id = $1 AND org_id = $2;

-- name: CountZonesByStore :one
SELECT COUNT(*) FROM zones WHERE store_id = $1 AND org_id = $2;
