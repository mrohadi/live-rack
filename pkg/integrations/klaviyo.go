package integrations

import (
	"encoding/json"
	"fmt"
)

// Klaviyo is an outbound adapter: it pushes sale events to Klaviyo's Events API
// for marketing automation. Building the request is pure; the HTTP POST is done
// by the caller.
type Klaviyo struct{}

// NewKlaviyo builds a Klaviyo adapter.
func NewKlaviyo() Klaviyo { return Klaviyo{} }

func (Klaviyo) Kind() string { return "klaviyo" }

// klaviyoEndpoint is the Events API ingestion URL.
const klaviyoEndpoint = "https://a.klaviyo.com/api/events/"

// OutboundRequest is a ready-to-send HTTP request description. Pure output.
type OutboundRequest struct {
	URL     string
	Headers map[string]string
	Body    []byte
}

// BuildTrackEvent builds a Klaviyo "Placed Order" event request for a sale tied
// to a customer email. apiKey authenticates via the Klaviyo-API-Key scheme. Pure.
func (Klaviyo) BuildTrackEvent(apiKey, email string, sale Sale) (OutboundRequest, error) {
	if apiKey == "" || email == "" {
		return OutboundRequest{}, fmt.Errorf("klaviyo: api key and email required")
	}
	payload := map[string]any{
		"data": map[string]any{
			"type": "event",
			"attributes": map[string]any{
				"metric":     map[string]any{"name": "Placed Order"},
				"profile":    map[string]any{"email": email},
				"value":      float64(sale.AmountCents) / 100,
				"unique_id":  sale.OrderID,
				"properties": map[string]any{"sku": sale.SKU, "qty": sale.Qty, "channel": sale.Channel},
				"time":       sale.OccurredAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return OutboundRequest{}, fmt.Errorf("klaviyo: marshal event: %w", err)
	}
	return OutboundRequest{
		URL: klaviyoEndpoint,
		Headers: map[string]string{
			"Authorization": "Klaviyo-API-Key " + apiKey,
			"Content-Type":  "application/json",
			"revision":      "2024-10-15",
		},
		Body: body,
	}, nil
}
