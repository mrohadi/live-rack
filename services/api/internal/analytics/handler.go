// Package analytics serves read-only ClickHouse-backed analytics for the
// Analytics screen: a 7x24 scan heatmap and per-zone performance. Reads are
// always scoped to the caller's org_id.
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
)

// Reader runs a ClickHouse query and returns the raw response body.
// *chstore.Client satisfies it.
type Reader interface {
	Query(ctx context.Context, sql string) ([]byte, error)
}

// Handler serves analytics endpoints.
type Handler struct {
	ch Reader
}

// New creates a Handler.
func New(ch Reader) *Handler {
	return &Handler{ch: ch}
}

// DisabledHandler registers the analytics routes returning 503 Service
// Unavailable. Used when CLICKHOUSE_URL is unset (MVP / demo deploy without
// ClickHouse). Lets the rest of the API stay up.
type DisabledHandler struct{}

// NewDisabled creates a no-op analytics handler.
func NewDisabled() *DisabledHandler {
	return &DisabledHandler{}
}

// Register mounts the disabled analytics endpoints.
func (h *DisabledHandler) Register(g *echo.Group) {
	disabled := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "analytics disabled (CLICKHOUSE_URL not set)")
	}
	g.GET("/analytics/heatmap", disabled)
	g.GET("/analytics/zones", disabled)
}

// Register mounts analytics routes on the authenticated API group.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/analytics/heatmap", h.Heatmap)
	g.GET("/analytics/zones", h.Zones)
}

// chResult is the envelope ClickHouse returns for FORMAT JSON.
type chResult[T any] struct {
	Data []T `json:"data"`
}

// parseCH unmarshals a ClickHouse FORMAT JSON body into rows. Pure.
func parseCH[T any](body []byte) ([]T, error) {
	var r chResult[T]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("analytics: parse clickhouse json: %w", err)
	}
	return r.Data, nil
}

const daysInWeek = 7
const hoursInDay = 24

type heatRow struct {
	Dow   int   `json:"dow"`  // ClickHouse toDayOfWeek: 1=Mon .. 7=Sun
	Hour  int   `json:"hour"` // 0..23
	Scans int64 `json:"scans"`
}

// HeatmapResponse is a dense 7x24 grid (row 0 = Monday) plus the peak cell for
// client-side colour scaling.
type HeatmapResponse struct {
	Grid [][]int64 `json:"grid"`
	Max  int64     `json:"max"`
}

// buildHeatGrid folds sparse (dow,hour,scans) rows into a dense 7x24 grid and
// reports the max cell. Out-of-range cells are ignored. Pure.
func buildHeatGrid(rows []heatRow) HeatmapResponse {
	grid := make([][]int64, daysInWeek)
	for i := range grid {
		grid[i] = make([]int64, hoursInDay)
	}
	var max int64
	for _, r := range rows {
		d := r.Dow - 1 // 1..7 -> 0..6
		if d < 0 || d >= daysInWeek || r.Hour < 0 || r.Hour >= hoursInDay {
			continue
		}
		grid[d][r.Hour] += r.Scans
		if grid[d][r.Hour] > max {
			max = grid[d][r.Hour]
		}
	}
	return HeatmapResponse{Grid: grid, Max: max}
}

// Heatmap godoc
//
//	@Summary	7x24 scan heatmap for the org (optionally one zone)
//	@Tags		analytics
//	@Produce	json
//	@Param		zone_id	query		string	false	"Filter to a single zone"
//	@Success	200		{object}	HeatmapResponse
//	@Router		/analytics/heatmap [get]
func (h *Handler) Heatmap(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	where := "org_id = '" + p.OrgID.String() + "'"
	if z := c.QueryParam("zone_id"); z != "" {
		zid, err := uuid.Parse(z)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid zone_id")
		}
		where += " AND zone_id = '" + zid.String() + "'"
	}

	sql := "SELECT toInt32(dow) AS dow, toInt32(hour) AS hour, toInt64(sum(scans)) AS scans " +
		"FROM heatmap_hourly WHERE " + where + " GROUP BY dow, hour ORDER BY dow, hour FORMAT JSON"

	body, err := h.ch.Query(c.Request().Context(), sql)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "heatmap query")
	}
	rows, err := parseCH[heatRow](body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "heatmap parse")
	}
	return c.JSON(http.StatusOK, buildHeatGrid(rows))
}

type zoneTotalRow struct {
	ZoneID  string `json:"zone_id"`
	Scans   int64  `json:"scans"`
	Picks   int64  `json:"picks"`
	Invalid int64  `json:"invalid"`
}

type zoneSeriesRow struct {
	ZoneID string `json:"zone_id"`
	Scans  int64  `json:"scans"`
}

// ZonePerf is one zone's rolled-up performance plus a recent hourly sparkline.
type ZonePerf struct {
	ZoneID  string  `json:"zone_id"`
	Scans   int64   `json:"scans"`
	Picks   int64   `json:"picks"`
	Invalid int64   `json:"invalid"`
	Spark   []int64 `json:"spark"`
}

// ZonesResponse powers the zone-performance bars on the Analytics screen.
type ZonesResponse struct {
	Zones []ZonePerf `json:"zones"`
}

// buildZonePerf joins per-zone totals (ordered, e.g. by scans desc) with their
// recent hourly series into sparkline arrays. Order follows totals. Pure.
func buildZonePerf(totals []zoneTotalRow, series []zoneSeriesRow) []ZonePerf {
	spark := make(map[string][]int64, len(totals))
	for _, s := range series {
		spark[s.ZoneID] = append(spark[s.ZoneID], s.Scans)
	}
	out := make([]ZonePerf, 0, len(totals))
	for _, t := range totals {
		sp := spark[t.ZoneID]
		if sp == nil {
			sp = []int64{}
		}
		out = append(out, ZonePerf{
			ZoneID: t.ZoneID, Scans: t.Scans, Picks: t.Picks, Invalid: t.Invalid, Spark: sp,
		})
	}
	return out
}

// Zones godoc
//
//	@Summary	Per-zone scan performance with recent hourly sparklines
//	@Tags		analytics
//	@Produce	json
//	@Success	200	{object}	ZonesResponse
//	@Router		/analytics/zones [get]
func (h *Handler) Zones(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	ctx := c.Request().Context()
	org := "org_id = '" + p.OrgID.String() + "'"

	totalsBody, err := h.ch.Query(ctx, "SELECT toString(zone_id) AS zone_id, "+
		"toInt64(sum(scans)) AS scans, toInt64(sum(picks)) AS picks, toInt64(sum(invalid)) AS invalid "+
		"FROM zone_perf_5m WHERE "+org+" GROUP BY zone_id ORDER BY scans DESC LIMIT 50 FORMAT JSON")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "zones query")
	}
	totals, err := parseCH[zoneTotalRow](totalsBody)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "zones parse")
	}

	seriesBody, err := h.ch.Query(ctx, "SELECT toString(zone_id) AS zone_id, "+
		"toInt64(sum(scans)) AS scans FROM zone_perf_5m WHERE "+org+
		" AND bucket >= now() - INTERVAL 24 HOUR GROUP BY zone_id, toStartOfHour(bucket) AS hb "+
		"ORDER BY zone_id, hb FORMAT JSON")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "zone series query")
	}
	series, err := parseCH[zoneSeriesRow](seriesBody)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "zone series parse")
	}

	return c.JSON(http.StatusOK, ZonesResponse{Zones: buildZonePerf(totals, series)})
}
