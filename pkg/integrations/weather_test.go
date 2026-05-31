package integrations_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

func TestOpenWeather_BuildURL(t *testing.T) {
	got := integrations.NewOpenWeather().BuildURL(40.7128, -74.006, "KEY123")
	u, err := url.Parse(got)
	require.NoError(t, err)
	q := u.Query()
	assert.Equal(t, "40.7128", q.Get("lat"))
	assert.Equal(t, "-74.006", q.Get("lon"))
	assert.Equal(t, "metric", q.Get("units"))
	assert.Equal(t, "KEY123", q.Get("appid"))
}

func TestOpenWeather_ParseCurrent(t *testing.T) {
	body := []byte(`{
		"name":"New York","dt":1717200000,
		"main":{"temp":18.5},
		"wind":{"speed":5},
		"weather":[{"main":"Rain"}]
	}`)
	got, err := integrations.NewOpenWeather().ParseCurrent(body)
	require.NoError(t, err)
	assert.Equal(t, "openweather", got.Source)
	assert.Equal(t, "New York", got.City)
	assert.Equal(t, 18.5, got.TempC)
	assert.Equal(t, "Rain", got.Condition)
	assert.InDelta(t, 18.0, got.WindKph, 0.001) // 5 m/s -> 18 kph
	assert.Equal(t, int64(1717200000), got.ObservedAt.Unix())
}

func TestOpenWeather_ParseCurrent_Invalid(t *testing.T) {
	_, err := integrations.NewOpenWeather().ParseCurrent([]byte("{bad"))
	assert.Error(t, err)
}

func TestOpenWeather_ParseCurrent_NoConditions(t *testing.T) {
	got, err := integrations.NewOpenWeather().ParseCurrent([]byte(`{"name":"X","main":{"temp":1}}`))
	require.NoError(t, err)
	assert.Equal(t, "", got.Condition)
}
