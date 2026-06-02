package poller_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/signals/internal/poller"
)

type fakeLister struct {
	stores []poller.StoreGeo
	err    error
}

func (f fakeLister) ListStoreGeo(context.Context) ([]poller.StoreGeo, error) {
	return f.stores, f.err
}

type fakeFetcher struct {
	body    []byte
	gotURLs []string
	err     error
}

func (f *fakeFetcher) Get(_ context.Context, url string) ([]byte, error) {
	f.gotURLs = append(f.gotURLs, url)
	return f.body, f.err
}

type capturePublisher struct {
	subjects []string
	values   []any
}

func (c *capturePublisher) Publish(_ context.Context, subject string, v any) error {
	c.subjects = append(c.subjects, subject)
	c.values = append(c.values, v)
	return nil
}

const owBody = `{"name":"NYC","dt":1717200000,"main":{"temp":18.5},"wind":{"speed":5},"weather":[{"main":"Rain"}]}`

func TestWeatherPoller_PublishesPerStore(t *testing.T) {
	org, store := uuid.New(), uuid.New()
	lister := fakeLister{stores: []poller.StoreGeo{
		{OrgID: org.String(), StoreID: store.String(), Lat: 40.7, Lon: -74.0},
	}}
	fetch := &fakeFetcher{body: []byte(owBody)}
	pub := &capturePublisher{}

	p := poller.NewWeatherPoller(lister, fetch, pub, "KEY")
	require.NoError(t, p.Poll(context.Background()))

	require.Len(t, pub.subjects, 1)
	assert.Equal(t, events.WeatherSignalSubject(org), pub.subjects[0])
	sig := pub.values[0].(events.WeatherSignal)
	assert.Equal(t, org, sig.OrgID)
	assert.Equal(t, store, sig.StoreID)
	assert.Equal(t, "Rain", sig.Condition)
	assert.InDelta(t, 18.0, sig.WindKph, 0.001)
	assert.Contains(t, fetch.gotURLs[0], "appid=KEY")
}

func TestWeatherPoller_ListError(t *testing.T) {
	p := poller.NewWeatherPoller(fakeLister{err: errors.New("db down")}, &fakeFetcher{}, &capturePublisher{}, "K")
	assert.Error(t, p.Poll(context.Background()))
}

func TestWeatherPoller_ContinuesPastBadStore(t *testing.T) {
	good := uuid.New()
	lister := fakeLister{stores: []poller.StoreGeo{
		{OrgID: "not-a-uuid", StoreID: uuid.New().String(), Lat: 1, Lon: 1},
		{OrgID: good.String(), StoreID: uuid.New().String(), Lat: 2, Lon: 2},
	}}
	pub := &capturePublisher{}
	p := poller.NewWeatherPoller(lister, &fakeFetcher{body: []byte(owBody)}, pub, "K")

	err := p.Poll(context.Background())
	assert.Error(t, err)           // first store failed
	assert.Len(t, pub.subjects, 1) // second store still published
	assert.Equal(t, events.WeatherSignalSubject(good), pub.subjects[0])
}
