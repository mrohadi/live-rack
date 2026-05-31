// Package insight is the rule-based recommender. It maps external signals
// (weather, transit) to operator Suggestions. It is pure: no IDs, clocks, or
// I/O — the insight service stamps those when building recommendation events.
package insight

import (
	"strings"

	"github.com/live-rack/pkg/events"
)

// Suggestion is a recommended operator action derived from a signal.
type Suggestion struct {
	Kind          string // "weather" | "transit"
	Severity      string // "info" | "action"
	Title         string
	Rationale     string
	SuggestedTask string
}

// Thresholds at which weather/transit rules fire.
const (
	HotTempC          = 28.0
	ColdTempC         = 5.0
	CommuterRushCount = 10
)

// wetConditions are OpenWeather "main" values that imply precipitation.
var wetConditions = map[string]bool{
	"Rain": true, "Drizzle": true, "Thunderstorm": true, "Snow": true,
}

// FromWeather returns suggestions for a weather signal. Pure.
func FromWeather(s events.WeatherSignal) []Suggestion {
	var out []Suggestion
	if wetConditions[s.Condition] {
		out = append(out, Suggestion{
			Kind: "weather", Severity: "action",
			Title:         "Move rain gear to entrance",
			Rationale:     strings.TrimSpace(s.Condition + " in " + s.City),
			SuggestedTask: "Front-of-store: stock umbrellas and rain gear",
		})
	}
	switch {
	case s.TempC >= HotTempC:
		out = append(out, Suggestion{
			Kind: "weather", Severity: "action",
			Title:         "Feature cold drinks up front",
			Rationale:     "High temperature drives cold-beverage demand",
			SuggestedTask: "Set up chilled-drinks display near entrance",
		})
	case s.TempC <= ColdTempC:
		out = append(out, Suggestion{
			Kind: "weather", Severity: "action",
			Title:         "Promote warm apparel",
			Rationale:     "Cold weather lifts outerwear sales",
			SuggestedTask: "Move warm apparel to a front zone",
		})
	}
	return out
}

// FromTransit returns suggestions for a transit signal. Pure.
func FromTransit(s events.TransitSignal) []Suggestion {
	if s.ArrivalsNext30m >= CommuterRushCount {
		return []Suggestion{{
			Kind: "transit", Severity: "action",
			Title:         "Staff up checkout for commuter rush",
			Rationale:     "Heavy nearby transit arrivals in the next 30 minutes",
			SuggestedTask: "Add a cashier to handle commuter rush",
		}}
	}
	return nil
}
