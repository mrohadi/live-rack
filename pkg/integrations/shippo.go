package integrations

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// headerShippoToken carries a shared secret Shippo echoes back per webhook
// (Shippo does not HMAC-sign payloads, so we authenticate by a configured token).
const headerShippoToken = "X-Shippo-Token" //nolint:gosec // G101: HTTP header name, not a credential.

// ShipmentUpdate is the canonical outbound-shipment tracking reading.
type ShipmentUpdate struct {
	Source         string    `json:"source"`
	Carrier        string    `json:"carrier"`
	TrackingNumber string    `json:"tracking_number"`
	Status         string    `json:"status"`
	ETA            time.Time `json:"eta"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Shippo adapts Shippo track_updated webhooks.
type Shippo struct{}

// NewShippo builds a Shippo adapter.
func NewShippo() Shippo { return Shippo{} }

func (Shippo) Kind() string { return "shippo" }

// Verify checks the shared token Shippo echoes in the X-Shippo-Token header.
func (Shippo) Verify(secret string, _ []byte, r *http.Request) error {
	got := r.Header.Get(headerShippoToken)
	if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
		return ErrInvalidSignature
	}
	return nil
}

type shippoEvent struct {
	Event string `json:"event"`
	Data  struct {
		Carrier        string `json:"carrier"`
		TrackingNumber string `json:"tracking_number"`
		ETA            string `json:"eta"`
		TrackingStatus struct {
			Status     string `json:"status"`
			StatusDate string `json:"status_date"`
		} `json:"tracking_status"`
	} `json:"data"`
}

func parseShippoTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

// ParseTracking normalises a Shippo track_updated payload into a ShipmentUpdate.
func (Shippo) ParseTracking(body []byte) (ShipmentUpdate, error) {
	var e shippoEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return ShipmentUpdate{}, fmt.Errorf("shippo: decode event: %w", err)
	}
	if e.Data.TrackingNumber == "" {
		return ShipmentUpdate{}, fmt.Errorf("shippo: event missing tracking_number")
	}
	return ShipmentUpdate{
		Source:         "shippo",
		Carrier:        e.Data.Carrier,
		TrackingNumber: e.Data.TrackingNumber,
		Status:         e.Data.TrackingStatus.Status,
		ETA:            parseShippoTime(e.Data.ETA),
		UpdatedAt:      parseShippoTime(e.Data.TrackingStatus.StatusDate),
	}, nil
}
