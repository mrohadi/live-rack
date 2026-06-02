-- name: AdjustItemLocationQty :one
INSERT INTO item_locations (org_id, store_id, zone_id, sku, qty)
VALUES (@org_id, @store_id, @zone_id, @sku, GREATEST(@qty::int, 0))
ON CONFLICT (org_id, zone_id, sku) DO UPDATE
SET qty = GREATEST(item_locations.qty + @qty::int, 0)
RETURNING *;

-- name: DecrementItemLocationQty :one
-- Guarded source decrement for a transfer: only succeeds when the location
-- holds at least @qty. Returns no rows (pgx.ErrNoRows) when stock is
-- insufficient or the location is missing, which the caller maps to 409.
UPDATE item_locations
SET qty = qty - @qty::int
WHERE org_id = @org_id AND zone_id = @zone_id AND sku = @sku
  AND qty >= @qty::int
RETURNING *;

-- name: SetItemLocationQty :one
-- Absolute on-hand correction for one zone (shrinkage, damage, cycle count).
-- Returns no rows when the location does not exist (caller maps to 404).
UPDATE item_locations
SET qty = GREATEST(@qty::int, 0)
WHERE org_id = @org_id AND store_id = @store_id AND zone_id = @zone_id AND sku = @sku
RETURNING *;

-- name: ListItemLocationsBySKU :many
-- Per-zone on-hand for a single SKU across a store (item detail drawer).
SELECT
    il.zone_id,
    il.qty,
    il.updated_at,
    COALESCE(z.name, '') AS zone_name
FROM item_locations il
LEFT JOIN zones z ON z.id = il.zone_id AND z.org_id = il.org_id
WHERE il.org_id = $1 AND il.store_id = $2 AND il.sku = $3
ORDER BY il.qty DESC;

-- name: ListInventoryByStore :many
SELECT
    il.id,
    il.org_id,
    il.store_id,
    il.zone_id,
    il.sku,
    il.qty,
    il.updated_at,
    COALESCE(i.name, '')      AS name,
    COALESCE(i.category, '')  AS category,
    COALESCE(i.status, '')    AS status,
    COALESCE(i.reorder_point, 0)::int AS reorder_point,
    COALESCE(i.price_cents, 0)::int   AS price_cents,
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