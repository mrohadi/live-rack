-- name: CreateCycleCount :one
INSERT INTO cycle_counts (org_id, store_id, zone_id, created_by)
VALUES (@org_id, @store_id, @zone_id, @created_by)
RETURNING *;

-- name: SnapshotCountLines :exec
-- Seed one line per SKU currently on-hand in the zone, capturing system_qty.
INSERT INTO cycle_count_lines (count_id, org_id, sku, system_qty)
SELECT @count_id, il.org_id, il.sku, il.qty
FROM item_locations il
WHERE il.org_id = @org_id AND il.store_id = @store_id AND il.zone_id = @zone_id;

-- name: GetCycleCount :one
SELECT * FROM cycle_counts
WHERE org_id = $1 AND id = $2;

-- name: ListCountLines :many
SELECT * FROM cycle_count_lines
WHERE count_id = $1
ORDER BY sku;

-- name: SetCountedQty :one
UPDATE cycle_count_lines
SET counted_qty = @counted_qty::int
WHERE org_id = @org_id AND count_id = @count_id AND sku = @sku
RETURNING *;

-- name: CompleteCycleCount :one
UPDATE cycle_counts
SET status = 'completed', completed_at = NOW()
WHERE org_id = @org_id AND id = @id AND status = 'open'
RETURNING *;

-- name: ListOpenCycleCountsByStore :many
SELECT * FROM cycle_counts
WHERE org_id = $1 AND store_id = $2 AND status = 'open'
ORDER BY created_at DESC;
