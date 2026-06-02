package domain

import (
	"math"

	"github.com/google/uuid"
)

// PickStop is one location a picker must visit. X,Y are the zone centroid on the
// store map; SKU + QtyRequested describe what to pick there.
type PickStop struct {
	ZoneID       uuid.UUID `json:"zone_id"`
	SKU          string    `json:"sku"`
	QtyRequested int       `json:"qty_requested"`
	X            float64   `json:"x"`
	Y            float64   `json:"y"`
}

// OptimizePickRoute orders stops into a short walking path using nearest-neighbour
// from the origin (the dock/staging point, typically 0,0). Deterministic: ties
// break by original slice order. Returns a new ordered slice; input is not mutated.
func OptimizePickRoute(stops []PickStop, originX, originY float64) []PickStop {
	n := len(stops)
	if n <= 1 {
		out := make([]PickStop, n)
		copy(out, stops)
		return out
	}

	remaining := make([]PickStop, n)
	copy(remaining, stops)

	out := make([]PickStop, 0, n)
	curX, curY := originX, originY
	for len(remaining) > 0 {
		best := 0
		bestD := dist(curX, curY, remaining[0].X, remaining[0].Y)
		for i := 1; i < len(remaining); i++ {
			if d := dist(curX, curY, remaining[i].X, remaining[i].Y); d < bestD {
				best, bestD = i, d
			}
		}
		chosen := remaining[best]
		out = append(out, chosen)
		curX, curY = chosen.X, chosen.Y
		remaining = append(remaining[:best], remaining[best+1:]...)
	}
	return out
}

// RouteLength returns the total walking distance of an ordered route, starting
// from the origin and visiting each stop in order.
func RouteLength(stops []PickStop, originX, originY float64) float64 {
	total := 0.0
	curX, curY := originX, originY
	for _, s := range stops {
		total += dist(curX, curY, s.X, s.Y)
		curX, curY = s.X, s.Y
	}
	return total
}

func dist(ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	return math.Sqrt(dx*dx + dy*dy)
}
