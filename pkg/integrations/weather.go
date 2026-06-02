package integrations

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// WeatherSignal is the canonical current-conditions reading for one store's geo,
// normalised across weather providers. Temperatures are Celsius.
type WeatherSignal struct {
	Source     string    `json:"source"`
	City       string    `json:"city"`
	TempC      float64   `json:"temp_c"`
	Condition  string    `json:"condition"`
	WindKph    float64   `json:"wind_kph"`
	ObservedAt time.Time `json:"observed_at"`
}

// OpenWeather builds requests and parses responses for the OpenWeather
// "current weather" API. Pure: no network calls live here.
type OpenWeather struct{}

// NewOpenWeather builds an OpenWeather adapter.
func NewOpenWeather() OpenWeather { return OpenWeather{} }

// Kind returns the signal-source identifier.
func (OpenWeather) Kind() string { return "openweather" }

const openWeatherBase = "https://api.openweathermap.org/data/2.5/weather"

// BuildURL returns the current-weather request URL for a store's coordinates.
// Units are metric so temperatures arrive in Celsius. Pure.
func (OpenWeather) BuildURL(lat, lon float64, apiKey string) string {
	q := url.Values{}
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	q.Set("units", "metric")
	q.Set("appid", apiKey)
	return openWeatherBase + "?" + q.Encode()
}

// owResponse is the subset of the OpenWeather payload we consume.
type owResponse struct {
	Name string `json:"name"`
	Dt   int64  `json:"dt"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"` // metric units: m/s
	} `json:"wind"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}

// msToKph converts metres/second to kilometres/hour. Pure.
func msToKph(ms float64) float64 { return ms * 3.6 }

// ParseCurrent normalises a verified OpenWeather body into a WeatherSignal. Pure.
func (o OpenWeather) ParseCurrent(body []byte) (WeatherSignal, error) {
	var r owResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return WeatherSignal{}, fmt.Errorf("integrations: parse openweather: %w", err)
	}
	cond := ""
	if len(r.Weather) > 0 {
		cond = r.Weather[0].Main
	}
	return WeatherSignal{
		Source:     o.Kind(),
		City:       r.Name,
		TempC:      r.Main.Temp,
		Condition:  cond,
		WindKph:    msToKph(r.Wind.Speed),
		ObservedAt: time.Unix(r.Dt, 0).UTC(),
	}, nil
}
