package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WeatherSignal is published by the signals worker after polling a weather
// provider for one store's geo. The insight engine consumes it to emit
// zone-move recommendations.
type WeatherSignal struct {
	OrgID      uuid.UUID `json:"org_id"`
	StoreID    uuid.UUID `json:"store_id"`
	Source     string    `json:"source"`
	City       string    `json:"city"`
	TempC      float64   `json:"temp_c"`
	Condition  string    `json:"condition"`
	WindKph    float64   `json:"wind_kph"`
	ObservedAt time.Time `json:"observed_at"`
}

const subjectWeatherSignal = "lr.%s.signal.weather"

// WeatherSignalSubject returns the per-org signal.weather subject.
func WeatherSignalSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectWeatherSignal, orgID)
}
