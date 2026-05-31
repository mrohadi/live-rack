package jobs

import "time"

// CombosLift recomputes co-purchase lift for SKU pairs that appear in the same
// order on one day. lift = P(a&b) / (P(a)*P(b)) = pair_orders*N / (support_a*support_b),
// where N is the day's order count. Writes into combos_lift (ReplacingMergeTree).
type CombosLift struct{}

// Name implements Job.
func (CombosLift) Name() string { return "combos_lift" }

// SQL implements Job.
func (CombosLift) SQL(day time.Time) string {
	d := DayString(day)
	return `INSERT INTO combos_lift
WITH order_skus AS (
    SELECT org_id, order_id, sku
    FROM sales_events_raw
    WHERE toDate(ts) = toDate('` + d + `') AND order_id != ''
    GROUP BY org_id, order_id, sku
)
SELECT
    pair.org_id,
    toDate('` + d + `') AS day,
    pair.sku_a,
    pair.sku_b,
    pair.pair_orders,
    pair.pair_orders * t.n / (sa.s * sb.s) AS lift
FROM (
    SELECT a.org_id AS org_id, a.sku AS sku_a, b.sku AS sku_b, uniqExact(a.order_id) AS pair_orders
    FROM order_skus AS a
    INNER JOIN order_skus AS b ON a.org_id = b.org_id AND a.order_id = b.order_id
    WHERE a.sku < b.sku
    GROUP BY a.org_id, a.sku, b.sku
) AS pair
INNER JOIN (SELECT org_id, uniqExact(order_id) AS n FROM order_skus GROUP BY org_id) AS t
    ON t.org_id = pair.org_id
INNER JOIN (SELECT org_id, sku, uniqExact(order_id) AS s FROM order_skus GROUP BY org_id, sku) AS sa
    ON sa.org_id = pair.org_id AND sa.sku = pair.sku_a
INNER JOIN (SELECT org_id, sku, uniqExact(order_id) AS s FROM order_skus GROUP BY org_id, sku) AS sb
    ON sb.org_id = pair.org_id AND sb.sku = pair.sku_b`
}
