-- name: CreateSaleEvent :one
INSERT INTO sales_events (ts, org_id, store_id, source, order_id, sku, qty, amount_cents, currency, channel)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: SalesSummary :one
SELECT
    COALESCE(SUM(amount_cents), 0)::BIGINT      AS revenue_cents,
    COALESCE(SUM(qty), 0)::BIGINT               AS units,
    COUNT(DISTINCT order_id)                    AS orders
FROM sales_events
WHERE org_id = $1 AND ts >= $2;

-- name: SalesByDay :many
SELECT
    time_bucket('1 day', ts) AS day,
    COALESCE(SUM(amount_cents), 0)::BIGINT AS revenue_cents
FROM sales_events
WHERE org_id = $1 AND ts >= $2
GROUP BY day
ORDER BY day;
