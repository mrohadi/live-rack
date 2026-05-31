package jobs

import "time"

// TimeToSell recomputes the average hours between an item's first place scan and
// its sale, per SKU, for one day. Writes into the time_to_sell ReplacingMergeTree
// (idempotent per day).
type TimeToSell struct{}

// Name implements Job.
func (TimeToSell) Name() string { return "time_to_sell" }

// SQL implements Job.
func (TimeToSell) SQL(day time.Time) string {
	d := DayString(day)
	return `INSERT INTO time_to_sell
SELECT
    s.org_id                                      AS org_id,
    toDate('` + d + `')                           AS day,
    s.sku                                         AS sku,
    avg(dateDiff('hour', p.first_place, s.ts))    AS avg_hours,
    count()                                       AS samples
FROM sales_events_raw AS s
INNER JOIN (
    SELECT org_id, sku, min(ts) AS first_place
    FROM scan_events_raw
    WHERE action = 'place' AND ts <= toDateTime(toDate('` + d + `') + 1)
    GROUP BY org_id, sku
) AS p ON s.org_id = p.org_id AND s.sku = p.sku
WHERE toDate(s.ts) = toDate('` + d + `') AND s.ts >= p.first_place
GROUP BY s.org_id, s.sku`
}

// SellThrough recomputes placed vs sold units per SKU for one day and the
// resulting sell-through rate. Writes into the sell_through ReplacingMergeTree.
type SellThrough struct{}

// Name implements Job.
func (SellThrough) Name() string { return "sell_through" }

// SQL implements Job.
func (SellThrough) SQL(day time.Time) string {
	d := DayString(day)
	return `INSERT INTO sell_through
SELECT
    org_id,
    toDate('` + d + `') AS day,
    sku,
    sumIf(cnt, kind = 'placed') AS placed,
    sumIf(cnt, kind = 'sold')   AS sold,
    if(placed > 0, sold / placed, 0) AS rate
FROM (
    SELECT org_id, sku, toUInt64(count()) AS cnt, 'placed' AS kind
    FROM scan_events_raw
    WHERE action = 'place' AND toDate(ts) = toDate('` + d + `')
    GROUP BY org_id, sku
    UNION ALL
    SELECT org_id, sku, toUInt64(sum(qty)) AS cnt, 'sold' AS kind
    FROM sales_events_raw
    WHERE toDate(ts) = toDate('` + d + `')
    GROUP BY org_id, sku
)
GROUP BY org_id, sku`
}
