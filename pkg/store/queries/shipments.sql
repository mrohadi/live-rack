-- name: CreateShipment :one
INSERT INTO shipments (org_id, store_id, pick_list_id, reference, created_by, status)
VALUES (@org_id, @store_id, @pick_list_id, @reference, @created_by, 'packing')
RETURNING *;

-- name: AddShipmentItem :one
INSERT INTO shipment_items (shipment_id, org_id, sku, qty)
VALUES (@shipment_id, @org_id, @sku, @qty::int)
RETURNING *;

-- name: GetShipment :one
SELECT * FROM shipments
WHERE org_id = $1 AND id = $2;

-- name: ListShipmentsByStore :many
SELECT
    s.*,
    (SELECT count(*) FROM shipment_items i WHERE i.shipment_id = s.id)::int AS item_count
FROM shipments s
WHERE s.org_id = $1 AND s.store_id = $2
ORDER BY s.created_at DESC;

-- name: ListShipmentItems :many
SELECT id, sku, qty FROM shipment_items
WHERE shipment_id = $1
ORDER BY sku;

-- name: MarkShipmentPacked :one
UPDATE shipments
SET status = 'packed'
WHERE org_id = @org_id AND id = @id AND status = 'packing'
RETURNING *;

-- name: MarkShipmentDispatched :one
UPDATE shipments
SET status = 'dispatched', carrier = @carrier, tracking_number = @tracking_number,
    dispatched_at = NOW()
WHERE org_id = @org_id AND id = @id AND status = 'packed'
RETURNING *;

-- name: CancelShipment :one
UPDATE shipments
SET status = 'cancelled'
WHERE org_id = @org_id AND id = @id AND status IN ('packing', 'packed')
RETURNING *;
