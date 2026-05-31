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

// Register mounts analytics routes on the authenticated API group.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/analytics/heatmap", h.Heatmap)
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
