// Package poller pulls external signals (weather, transit) per store on a
// schedule and publishes them onto NATS for the insight engine. The provider
// adapters (URL building, response parsing) are pure and live in pkg/integrations;
// this package owns the I/O: store geo lookup, HTTP fetch, and publishing.
package poller

import (
	"context"
	"fmt"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/integrations"
)

// StoreGeo is one store's coordinates for geo-scoped signal polling.
type StoreGeo struct {
	OrgID   string
	StoreID string
	Lat     float64
	Lon     float64
}

// StoreLister returns the stores to poll. Implementations read Postgres.
type StoreLister interface {
	ListStoreGeo(ctx context.Context) ([]StoreGeo, error)
}

// Fetcher performs the outbound HTTP GET. Injected so the poller is testable
// without network access.
type Fetcher interface {
	Get(ctx context.Context, url string) ([]byte, error)
}

// WeatherPoller fetches current weather for each store and publishes a
// WeatherSignal per store.
type WeatherPoller struct {
	stores StoreLister
	fetch  Fetcher
	pub    events.Publisher
	ow     integrations.OpenWeather
	apiKey string
}

// NewWeatherPoller builds a WeatherPoller.
func NewWeatherPoller(stores StoreLister, fetch Fetcher, pub events.Publisher, apiKey string) *WeatherPoller {
	return &WeatherPoller{stores: stores, fetch: fetch, pub: pub, ow: integrations.NewOpenWeather(), apiKey: apiKey}
}

// Poll fetches and publishes weather for every store. A per-store failure is
// returned wrapped after attempting the remaining stores.
func (p *WeatherPoller) Poll(ctx context.Context) error {
	stores, err := p.stores.ListStoreGeo(ctx)
	if err != nil {
		return fmt.Errorf("poller: list stores: %w", err)
	}
	var firstErr error
	for _, s := range stores {
		if err := p.pollStore(ctx, s); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (p *WeatherPoller) pollStore(ctx context.Context, s StoreGeo) error {
	orgID, err := parseUUID(s.OrgID)
	if err != nil {
		return fmt.Errorf("poller: store %s org id: %w", s.StoreID, err)
	}
	storeID, err := parseUUID(s.StoreID)
	if err != nil {
		return fmt.Errorf("poller: store id %s: %w", s.StoreID, err)
	}

	body, err := p.fetch.Get(ctx, p.ow.BuildURL(s.Lat, s.Lon, p.apiKey))
	if err != nil {
		return fmt.Errorf("poller: fetch weather for %s: %w", s.StoreID, err)
	}
	sig, err := p.ow.ParseCurrent(body)
	if err != nil {
		return fmt.Errorf("poller: parse weather for %s: %w", s.StoreID, err)
	}

	evt := events.WeatherSignal{
		OrgID: orgID, StoreID: storeID, Source: sig.Source, City: sig.City,
		TempC: sig.TempC, Condition: sig.Condition, WindKph: sig.WindKph, ObservedAt: sig.ObservedAt,
	}
	if err := p.pub.Publish(ctx, events.WeatherSignalSubject(orgID), evt); err != nil {
		return fmt.Errorf("poller: publish weather for %s: %w", s.StoreID, err)
	}
	return nil
}
