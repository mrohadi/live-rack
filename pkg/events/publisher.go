package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// Publisher emits domain events. Implementations must be safe for concurrent use.
type Publisher interface {
	Publish(ctx context.Context, subject string, v any) error
}

// NATSPublisher publishes JSON payloads onto a JetStream subject.
type NATSPublisher struct {
	js nats.JetStreamContext
}

// NewNATSPublisher wraps a JetStream context.
func NewNATSPublisher(js nats.JetStreamContext) *NATSPublisher {
	return &NATSPublisher{js: js}
}

// Publish marshals v to JSON and async-publishes it on subject.
func (p *NATSPublisher) Publish(_ context.Context, subject string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("events.Publish marshal: %w", err)
	}
	if _, err := p.js.PublishAsync(subject, b); err != nil {
		return fmt.Errorf("events.Publish %s: %w", subject, err)
	}
	return nil
}
