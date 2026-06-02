package poller

import (
	"context"
	"fmt"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/integrations"
)

const defaultTransitRadiusM = 500

// TransitPoller fetches nearby transit for each store and publishes a
// TransitSignal per store.
type TransitPoller struct {
	stores  StoreLister
	fetch   Fetcher
	pub     events.Publisher
	tr      integrations.Transit
	apiKey  string
	radiusM int
}

// NewTransitPoller builds a TransitPoller.
func NewTransitPoller(stores StoreLister, fetch Fetcher, pub events.Publisher, apiKey string) *TransitPoller {
	return &TransitPoller{stores: stores, fetch: fetch, pub: pub, tr: integrations.NewTransit(), apiKey: apiKey, radiusM: defaultTransitRadiusM}
}

// Poll fetches and publishes transit for every store. A per-store failure is
// returned wrapped after attempting the remaining stores.
func (p *TransitPoller) Poll(ctx context.Context) error {
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

func (p *TransitPoller) pollStore(ctx context.Context, s StoreGeo) error {
	orgID, err := parseUUID(s.OrgID)
	if err != nil {
		return fmt.Errorf("poller: store %s org id: %w", s.StoreID, err)
	}
	storeID, err := parseUUID(s.StoreID)
	if err != nil {
		return fmt.Errorf("poller: store id %s: %w", s.StoreID, err)
	}

	body, err := p.fetch.Get(ctx, p.tr.BuildURL(s.Lat, s.Lon, p.radiusM, p.apiKey))
	if err != nil {
		return fmt.Errorf("poller: fetch transit for %s: %w", s.StoreID, err)
	}
	sig, err := p.tr.ParseNearby(body)
	if err != nil {
		return fmt.Errorf("poller: parse transit for %s: %w", s.StoreID, err)
	}

	evt := events.TransitSignal{
		OrgID: orgID, StoreID: storeID, Source: sig.Source, StopCount: sig.StopCount,
		ArrivalsNext30m: sig.ArrivalsNext30m, BusiestRoute: sig.BusiestRoute, ObservedAt: sig.ObservedAt,
	}
	if err := p.pub.Publish(ctx, events.TransitSignalSubject(orgID), evt); err != nil {
		return fmt.Errorf("poller: publish transit for %s: %w", s.StoreID, err)
	}
	return nil
}
