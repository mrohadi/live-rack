package integrations

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// TransitSignal is the canonical nearby-transit reading for one store's geo,
// normalised across transit providers.
type TransitSignal struct {
	Source          string    `json:"source"`
	StopCount       int       `json:"stop_count"`
	ArrivalsNext30m int       `json:"arrivals_next_30m"`
	BusiestRoute    string    `json:"busiest_route"`
	ObservedAt      time.Time `json:"observed_at"`
}

// Transit builds requests and parses responses for a generic nearby-transit
// API. Pure: no network calls live here.
type Transit struct{}

// NewTransit builds a Transit adapter.
func NewTransit() Transit { return Transit{} }

// Kind returns the signal-source identifier.
func (Transit) Kind() string { return "transit" }

const transitBase = "https://transit.example.com/v1/nearby"

// BuildURL returns the nearby-arrivals request URL for a store's coordinates
// within radiusM metres. Pure.
func (Transit) BuildURL(lat, lon float64, radiusM int, apiKey string) string {
	q := url.Values{}
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	q.Set("radius", strconv.Itoa(radiusM))
	q.Set("api_key", apiKey)
	return transitBase + "?" + q.Encode()
}

type transitResponse struct {
	Stops    []struct{} `json:"stops"`
	Arrivals []struct {
		Route   string `json:"route"`
		Minutes int    `json:"minutes"`
	} `json:"arrivals"`
}

const arrivalsWindowMin = 30

// ParseNearby normalises a transit body into a TransitSignal: counts stops,
// arrivals due within the next 30 minutes, and the most-frequent route. Pure.
func (tr Transit) ParseNearby(body []byte) (TransitSignal, error) {
	var r transitResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return TransitSignal{}, fmt.Errorf("integrations: parse transit: %w", err)
	}

	counts := make(map[string]int)
	soon := 0
	for _, a := range r.Arrivals {
		if a.Minutes <= arrivalsWindowMin {
			soon++
			counts[a.Route]++
		}
	}

	busiest, best := "", 0
	for route, n := range counts {
		// Tie-break deterministically by route name to keep output stable.
		if n > best || (n == best && route < busiest) {
			busiest, best = route, n
		}
	}

	return TransitSignal{
		Source:          tr.Kind(),
		StopCount:       len(r.Stops),
		ArrivalsNext30m: soon,
		BusiestRoute:    busiest,
		ObservedAt:      time.Now().UTC(),
	}, nil
}
