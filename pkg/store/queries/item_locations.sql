-- name: AdjustItemLocationQty :one
INSERT INTO item_locations (org_id, store_id, zone_id, sku, qty)
VALUES (@org_id, @store_id, @zone_id, @sku, GREATEST(@qty::int, 0))
ON CONFLICT (org_id, zone_id, sku) DO UPDATE
SET qty = GREATEST(item_locations.qty + @qty::int, 0)
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
    COALESCE(i.status, '')   AS status,
    COALESCE((
        SELECT count(*) FROM scan_events se
        WHERE se.org_id = il.org_id AND se.zone_id = il.zone_id AND se.sku = il.sku
          AND se.action = 'pick' AND se.valid
          AND se.ts >= NOW() - INTERVAL '7 days'
    ), 0)::int  AS picks_7d,
    COALESCE((
        SELECT count(*) FROM scan_events se
        WHERE se.org_id = il.org_id AND se.zone_id = il.zone_id AND se.sku = il.sku
          AND se.action = 'pick' AND se.valid
          AND se.ts >= NOW() - INTERVAL '30 days'
    ), 0)::int  AS picks_30d
FROM item_locations il
LEFT JOIN items i ON i.org_id = il.org_id AND i.sku = il.sku
WHERE il.org_id = $1 AND il.store_id = $2
ORDER BY il.updated_at DESC;