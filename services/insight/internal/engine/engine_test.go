package engine_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/insight/internal/engine"
)

type capturePub struct {
	subjects []string
	recs     []events.Recommendation
}

func (c *capturePub) Publish(_ context.Context, subject string, v any) error {
	c.subjects = append(c.subjects, subject)
	c.recs = append(c.recs, v.(events.Recommendation))
	return nil
}

func fixedEngine(pub events.Publisher) *engine.Engine {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ts := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	return engine.New(pub, func() uuid.UUID { return id }, func() time.Time { return ts })
}

func TestHandleWeather_EmitsRecommendations(t *testing.T) {
	org, store := uuid.New(), uuid.New()
	data, _ := json.Marshal(events.WeatherSignal{OrgID: org, StoreID: store, City: "NYC", Condition: "Rain", TempC: 30})

	pub := &capturePub{}
	require.NoError(t, fixedEngine(pub).HandleWeather(context.Background(), data))

	require.Len(t, pub.recs, 2) // rain + heat
	assert.Equal(t, events.RecommendationSubject(org), pub.subjects[0])
	assert.Equal(t, org, pub.recs[0].OrgID)
	assert.Equal(t, store, pub.recs[0].StoreID)
	assert.Equal(t, "weather", pub.recs[0].Kind)
	assert.NotEmpty(t, pub.recs[0].SuggestedTask)
}

func TestHandleTransit_Quiet_EmitsNothing(t *testing.T) {
	data, _ := json.Marshal(events.TransitSignal{OrgID: uuid.New(), ArrivalsNext30m: 2})
	pub := &capturePub{}
	require.NoError(t, fixedEngine(pub).HandleTransit(context.Background(), data))
	assert.Empty(t, pub.recs)
}

func TestHandleWeather_MissingOrg(t *testing.T) {
	data, _ := json.Marshal(events.WeatherSignal{Condition: "Rain"})
	pub := &capturePub{}
	assert.Error(t, fixedEngine(pub).HandleWeather(context.Background(), data))
}
