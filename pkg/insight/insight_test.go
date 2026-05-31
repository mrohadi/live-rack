package insight_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/insight"
)

func TestFromWeather_RainAndHeat(t *testing.T) {
	got := insight.FromWeather(events.WeatherSignal{City: "NYC", Condition: "Rain", TempC: 30})
	assert.Len(t, got, 2)
	assert.Equal(t, "Move rain gear to entrance", got[0].Title)
	assert.Equal(t, "Feature cold drinks up front", got[1].Title)
	for _, s := range got {
		assert.Equal(t, "action", s.Severity)
		assert.NotEmpty(t, s.SuggestedTask)
	}
}

func TestFromWeather_Cold(t *testing.T) {
	got := insight.FromWeather(events.WeatherSignal{Condition: "Clear", TempC: 2})
	assert.Len(t, got, 1)
	assert.Equal(t, "Promote warm apparel", got[0].Title)
}

func TestFromWeather_Mild_NoSuggestions(t *testing.T) {
	assert.Empty(t, insight.FromWeather(events.WeatherSignal{Condition: "Clouds", TempC: 18}))
}

func TestFromTransit_Rush(t *testing.T) {
	got := insight.FromTransit(events.TransitSignal{ArrivalsNext30m: 12})
	assert.Len(t, got, 1)
	assert.Equal(t, "transit", got[0].Kind)
}

func TestFromTransit_Quiet(t *testing.T) {
	assert.Empty(t, insight.FromTransit(events.TransitSignal{ArrivalsNext30m: 3}))
}
