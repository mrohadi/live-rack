-- name: AdjustItemLocationQty :one
INSERT INTO item_locations (org_id, store_id, zone_id, sku, qty)
VALUES ($1, $2, $3, $4, GREATEST($5, 0))
ON CONFLICT (org_id, zone_id, sku) DO UPDATE
SET qty = GREATEST(item_locations.qty + $5, 0)
RETURNING *;

-- name: ListInventoryByStore :many
SELECT
    il.id,
    il.org_id,
    il.store_id,
    il.zone_id,
    il.sku,
    il.qty,
    il.updated_at,
    COALESCE(i.name, '')     AS name,
    COALESCE(i.category, '') AS category,
    COALESCE(i.status, '')   AS status
FROM item_locations il
LEFT JOIN items i ON i.org_id = il.org_id AND i.sku = il.sku
WHERE il.org_id = $1 AND il.store_id = $2
ORDER BY il.updated_at DESC;
