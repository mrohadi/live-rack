// Package engine wires the pure insight rules to NATS: it decodes signal
// events, runs the recommender, and publishes recommendation.created events.
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/insight"
)

// Engine turns signals into recommendation events.
type Engine struct {
	pub   events.Publisher
	newID func() uuid.UUID
	now   func() time.Time
}

// New builds an Engine. idGen and clock are injectable for deterministic tests;
// pass nil to use uuid.New and time.Now.
func New(pub events.Publisher, idGen func() uuid.UUID, clock func() time.Time) *Engine {
	if idGen == nil {
		idGen = uuid.New
	}
	if clock == nil {
		clock = time.Now
	}
	return &Engine{pub: pub, newID: idGen, now: clock}
}

// HandleWeather decodes a weather signal and publishes any recommendations.
func (e *Engine) HandleWeather(ctx context.Context, data []byte) error {
	var s events.WeatherSignal
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("engine: decode weather signal: %w", err)
	}
	return e.emit(ctx, s.OrgID, s.StoreID, insight.FromWeather(s))
}

// HandleTransit decodes a transit signal and publishes any recommendations.
func (e *Engine) HandleTransit(ctx context.Context, data []byte) error {
	var s events.TransitSignal
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("engine: decode transit signal: %w", err)
	}
	return e.emit(ctx, s.OrgID, s.StoreID, insight.FromTransit(s))
}

func (e *Engine) emit(ctx context.Context, orgID, storeID uuid.UUID, suggestions []insight.Suggestion) error {
	if orgID == uuid.Nil {
		return fmt.Errorf("engine: signal missing org_id")
	}
	for _, s := range suggestions {
		rec := events.Recommendation{
			ID: e.newID(), OrgID: orgID, StoreID: storeID,
			Kind: s.Kind, Severity: s.Severity, Title: s.Title,
			Rationale: s.Rationale, SuggestedTask: s.SuggestedTask, CreatedAt: e.now().UTC(),
		}
		if err := e.pub.Publish(ctx, events.RecommendationSubject(orgID), rec); err != nil {
			return fmt.Errorf("engine: publish recommendation: %w", err)
		}
	}
	return nil
}
