package integrations_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

const shippoBody = `{"event":"track_updated","data":{
	"carrier":"usps","tracking_number":"9400111","eta":"2026-06-03T00:00:00Z",
	"tracking_status":{"status":"DELIVERED","status_date":"2026-06-02T15:04:05Z"}}}`

func shippoReq(token string) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	if token != "" {
		r.Header.Set("X-Shippo-Token", token)
	}
	return r
}

func TestShippo_Verify(t *testing.T) {
	s := integrations.NewShippo()
	require.NoError(t, s.Verify("sek", nil, shippoReq("sek")))
	assert.ErrorIs(t, s.Verify("sek", nil, shippoReq("nope")), integrations.ErrInvalidSignature)
	assert.ErrorIs(t, s.Verify("sek", nil, shippoReq("")), integrations.ErrInvalidSignature)
}

func TestShippo_ParseTracking(t *testing.T) {
	got, err := integrations.NewShippo().ParseTracking([]byte(shippoBody))
	require.NoError(t, err)
	assert.Equal(t, "shippo", got.Source)
	assert.Equal(t, "usps", got.Carrier)
	assert.Equal(t, "9400111", got.TrackingNumber)
	assert.Equal(t, "DELIVERED", got.Status)
	assert.Equal(t, 2026, got.ETA.Year())
	assert.Equal(t, 15, got.UpdatedAt.Hour())
}

func TestShippo_ParseTracking_MissingNumber(t *testing.T) {
	_, err := integrations.NewShippo().ParseTracking([]byte(`{"event":"track_updated","data":{}}`))
	assert.Error(t, err)
}
