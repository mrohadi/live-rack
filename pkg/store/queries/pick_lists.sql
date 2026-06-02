-- name: CreatePickList :one
INSERT INTO pick_lists (org_id, store_id, reference, created_by, assignee_id, status)
VALUES (@org_id, @store_id, @reference, @created_by, @assignee_id, 'open')
RETURNING *;

-- name: AddPickLine :one
INSERT INTO pick_list_lines (list_id, org_id, zone_id, sku, qty_requested, seq)
VALUES (@list_id, @org_id, @zone_id, @sku, @qty_requested::int, @seq::int)
RETURNING *;

-- name: GetPickList :one
SELECT * FROM pick_lists
WHERE org_id = $1 AND id = $2;

-- name: ListPickListsByStore :many
SELECT
    pl.*,
    (SELECT count(*) FROM pick_list_lines l WHERE l.list_id = pl.id)::int AS line_count,
    (SELECT count(*) FROM pick_list_lines l WHERE l.list_id = pl.id AND l.status <> 'pending')::int AS done_count
FROM pick_lists pl
WHERE pl.org_id = $1 AND pl.store_id = $2
ORDER BY pl.created_at DESC;

-- name: ListPickListLines :many
SELECT
    pl.id,
    pl.zone_id,
    pl.sku,
    pl.qty_requested,
    pl.qty_picked,
    pl.seq,
    pl.status,
    COALESCE(z.name, '')   AS zone_name,
    COALESCE(z.x, 0)::float8 AS zone_x,
    COALESCE(z.y, 0)::float8 AS zone_y
FROM pick_list_lines pl
LEFT JOIN zones z ON z.id = pl.zone_id AND z.org_id = pl.org_id
WHERE pl.list_id = $1
ORDER BY pl.seq;

-- name: SetPickLinePicked :one
UPDATE pick_list_lines
SET qty_picked = @qty_picked::int, status = @status::text
WHERE org_id = @org_id AND id = @id
RETURNING *;

-- name: StartPickList :one
UPDATE pick_lists
SET status = 'picking'
WHERE org_id = $1 AND id = $2 AND status = 'open'
RETURNING *;

-- name: CompletePickList :one
UPDATE pick_lists
SET status = 'completed', completed_at = NOW()
WHERE org_id = @org_id AND id = @id AND status IN ('open', 'picking')
RETURNING *;

-- name: CancelPickList :one
UPDATE pick_lists
SET status = 'cancelled'
WHERE org_id = @org_id AND id = @id AND status IN ('open', 'picking')
RETURNING *;

-- name: ResolvePickSource :one
-- Best source location for a SKU: the zone holding the most on-hand units, with
-- its map coordinates so the route can be optimised. Returns no rows when the
-- SKU has zero on-hand anywhere in the store.
SELECT
    il.zone_id,
    il.qty,
    COALESCE(z.x, 0)::float8  AS zone_x,
    COALESCE(z.y, 0)::float8  AS zone_y,
    COALESCE(z.name, '')      AS zone_name
FROM item_locations il
LEFT JOIN zones z ON z.id = il.zone_id AND z.org_id = il.org_id
WHERE il.org_id = $1 AND il.store_id = $2 AND il.sku = $3 AND il.qty > 0
ORDER BY il.qty DESC
LIMIT 1;
