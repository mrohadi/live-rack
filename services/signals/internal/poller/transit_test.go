package poller_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/signals/internal/poller"
)

const transitBody = `{"stops":[{},{}],"arrivals":[{"route":"M1","minutes":3},{"route":"M1","minutes":10},{"route":"M2","minutes":40}]}`

func TestTransitPoller_PublishesPerStore(t *testing.T) {
	org, store := uuid.New(), uuid.New()
	lister := fakeLister{stores: []poller.StoreGeo{
		{OrgID: org.String(), StoreID: store.String(), Lat: 40.7, Lon: -74.0},
	}}
	fetch := &fakeFetcher{body: []byte(transitBody)}
	pub := &capturePublisher{}

	p := poller.NewTransitPoller(lister, fetch, pub, "KEY")
	require.NoError(t, p.Poll(context.Background()))

	require.Len(t, pub.subjects, 1)
	assert.Equal(t, events.TransitSignalSubject(org), pub.subjects[0])
	sig := pub.values[0].(events.TransitSignal)
	assert.Equal(t, org, sig.OrgID)
	assert.Equal(t, 2, sig.StopCount)
	assert.Equal(t, 2, sig.ArrivalsNext30m)
	assert.Equal(t, "M1", sig.BusiestRoute)
	assert.Contains(t, fetch.gotURLs[0], "radius=500")
}
