package integrations_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

func TestTransit_BuildURL(t *testing.T) {
	got := integrations.NewTransit().BuildURL(40.7, -74.0, 500, "KEY")
	u, err := url.Parse(got)
	require.NoError(t, err)
	q := u.Query()
	assert.Equal(t, "40.7", q.Get("lat"))
	assert.Equal(t, "500", q.Get("radius"))
	assert.Equal(t, "KEY", q.Get("api_key"))
}

func TestTransit_ParseNearby(t *testing.T) {
	body := []byte(`{
		"stops":[{},{}],
		"arrivals":[
			{"route":"M1","minutes":3},
			{"route":"M1","minutes":12},
			{"route":"M2","minutes":8},
			{"route":"M3","minutes":45}
		]
	}`)
	got, err := integrations.NewTransit().ParseNearby(body)
	require.NoError(t, err)
	assert.Equal(t, "transit", got.Source)
	assert.Equal(t, 2, got.StopCount)
	assert.Equal(t, 3, got.ArrivalsNext30m) // M3@45 excluded
	assert.Equal(t, "M1", got.BusiestRoute) // 2 arrivals within window
}

func TestTransit_ParseNearby_Invalid(t *testing.T) {
	_, err := integrations.NewTransit().ParseNearby([]byte("{bad"))
	assert.Error(t, err)
}

func TestTransit_ParseNearby_Empty(t *testing.T) {
	got, err := integrations.NewTransit().ParseNearby([]byte(`{"stops":[],"arrivals":[]}`))
	require.NoError(t, err)
	assert.Equal(t, 0, got.ArrivalsNext30m)
	assert.Equal(t, "", got.BusiestRoute)
}
