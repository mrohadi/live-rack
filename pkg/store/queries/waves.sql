-- name: CreateWave :one
INSERT INTO waves (org_id, store_id, reference, created_by, status)
VALUES (@org_id, @store_id, @reference, @created_by, 'open')
RETURNING *;

-- name: AssignListsToWave :exec
-- Attach the given open pick lists to a wave (skips lists already in a wave).
UPDATE pick_lists
SET wave_id = @wave_id
WHERE org_id = @org_id AND store_id = @store_id
  AND id = ANY(@list_ids::uuid[]) AND wave_id IS NULL;

-- name: GetWave :one
SELECT * FROM waves
WHERE org_id = $1 AND id = $2;

-- name: ListWavesByStore :many
SELECT
    w.*,
    (SELECT count(*) FROM pick_lists pl WHERE pl.wave_id = w.id)::int AS list_count
FROM waves w
WHERE w.org_id = $1 AND w.store_id = $2
ORDER BY w.created_at DESC;

-- name: ListWaveMergedLines :many
-- Aggregate member order lines into one stop per SKU+zone, summing demand and
-- picked qty across the wave's orders. Only mapped stops are returned.
SELECT
    pll.sku,
    pll.zone_id,
    COALESCE(z.name, '')     AS zone_name,
    COALESCE(z.x, 0)::float8 AS zone_x,
    COALESCE(z.y, 0)::float8 AS zone_y,
    SUM(pll.qty_requested)::int AS qty_requested,
    SUM(pll.qty_picked)::int    AS qty_picked,
    COUNT(*)::int               AS order_count
FROM pick_list_lines pll
JOIN pick_lists pl ON pl.id = pll.list_id AND pl.org_id = pll.org_id
LEFT JOIN zones z ON z.id = pll.zone_id AND z.org_id = pll.org_id
WHERE pl.org_id = $1 AND pl.wave_id = $2 AND pll.zone_id IS NOT NULL
GROUP BY pll.sku, pll.zone_id, z.name, z.x, z.y
ORDER BY pll.sku;

-- name: ListWaveStopMemberLines :many
-- Member order lines for one SKU+zone stop, in FIFO order (oldest order first),
-- for allocating a merged picked quantity back to orders.
SELECT pll.id, pll.list_id, pll.qty_requested, pll.qty_picked
FROM pick_list_lines pll
JOIN pick_lists pl ON pl.id = pll.list_id AND pl.org_id = pll.org_id
WHERE pl.org_id = @org_id AND pl.wave_id = @wave_id
  AND pll.sku = @sku AND pll.zone_id = @zone_id
ORDER BY pl.created_at, pll.seq;

-- name: StartWave :one
UPDATE waves
SET status = 'picking'
WHERE org_id = $1 AND id = $2 AND status = 'open'
RETURNING *;

-- name: CompleteWave :one
UPDATE waves
SET status = 'completed', completed_at = NOW()
WHERE org_id = @org_id AND id = @id AND status IN ('open', 'picking')
RETURNING *;
