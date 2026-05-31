// Package integrations defines the uniform third-party adapter contract and the
// Shopify/Square POS connectors. Adapters are pure: they verify inbound webhook
// authenticity and parse vendor payloads into canonical Sale events. Persistence,
// HTTP routing, and idempotency live in the API gateway and ingest worker.
package integrations

import (
	"errors"
	"net/http"
	"time"
)

// ErrInvalidSignature is returned when a webhook fails authenticity verification.
var ErrInvalidSignature = errors.New("integrations: invalid webhook signature")

// Sale is the canonical sale line normalised across vendors. Amount is in minor
// units (cents) to avoid float rounding.
type Sale struct {
	Source      string    `json:"source"`
	OrderID     string    `json:"order_id"`
	SKU         string    `json:"sku"`
	Qty         int32     `json:"qty"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	Channel     string    `json:"channel"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// Adapter is the uniform connector contract. Implementations must be safe for
// concurrent use and free of network calls in Verify/ParseSales.
type Adapter interface {
	// Kind returns the connector identifier ("shopify", "square").
	Kind() string
	// EventID extracts the vendor's unique delivery id for idempotency, or "".
	EventID(headers http.Header) string
	// Verify checks the webhook signature against the shared secret.
	Verify(secret string, body []byte, headers http.Header) error
	// ParseSales normalises a verified webhook body into canonical sales.
	ParseSales(body []byte) ([]Sale, error)
}

// Registry maps connector kinds to adapters.
type Registry struct {
	byKind map[string]Adapter
}

// NewRegistry builds a registry from the given adapters.
func NewRegistry(adapters ...Adapter) *Registry {
	r := &Registry{byKind: make(map[string]Adapter, len(adapters))}
	for _, a := range adapters {
		r.byKind[a.Kind()] = a
	}
	return r
}

// Get returns the adapter for kind, or ok=false if unregistered.
func (r *Registry) Get(kind string) (Adapter, bool) {
	a, ok := r.byKind[kind]
	return a, ok
}
