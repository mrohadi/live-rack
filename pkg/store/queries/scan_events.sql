-- name: CreateScanEvent :one
INSERT INTO scan_events (ts, org_id, store_id, zone_id, scanner_id, sku, action, valid, reason)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetLastScanForSKU :one
SELECT * FROM scan_events
WHERE org_id = $1 AND zone_id = $2 AND sku = $3
ORDER BY ts DESC
LIMIT 1;

-- name: ListScanEventsByZone :many
SELECT * FROM scan_events
WHERE org_id = $1 AND zone_id = $2
ORDER BY ts DESC
LIMIT $3;

-- name: ListScanEventsBySKU :many
-- Recent scan timeline for one SKU across a store (item detail drawer).
SELECT * FROM scan_events
WHERE org_id = $1 AND store_id = $2 AND sku = $3
ORDER BY ts DESC
LIMIT $4;