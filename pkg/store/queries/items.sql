-- name: UpsertItem :one
INSERT INTO items (org_id, sku, name, category, status, reorder_point)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (org_id, sku) DO UPDATE
SET name = EXCLUDED.name, category = EXCLUDED.category,
    status = EXCLUDED.status, reorder_point = EXCLUDED.reorder_point
RETURNING *;

-- name: GetItemBySKU :one
SELECT * FROM items
WHERE org_id = $1 AND sku = $2;

-- name: ListItems :many
SELECT * FROM items
WHERE org_id = $1
ORDER BY name;

-- name: UpdateItem :one
-- Edit master-catalog fields for an existing SKU (LR-310).
UPDATE items
SET name = @name, category = @category, status = @status, reorder_point = @reorder_point::int
WHERE org_id = @org_id AND sku = @sku
RETURNING *;
