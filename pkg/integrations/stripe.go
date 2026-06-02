package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// headerStripeSignature carries the timestamped HMAC: "t=<ts>,v1=<hexmac>".
const headerStripeSignature = "Stripe-Signature"

// Stripe adapts Stripe (Connect) charge webhooks. The signature is
// HMAC-SHA256("<t>.<body>") hex-encoded, per Stripe's scheme.
type Stripe struct{}

// NewStripe builds a Stripe adapter.
func NewStripe() Stripe { return Stripe{} }

func (Stripe) Kind() string { return "stripe" }

// AccountHandle returns the connected account id (Stripe-Account header) used to
// route the webhook to an org.
func (Stripe) AccountHandle(_ []byte, r *http.Request) string {
	return r.Header.Get("Stripe-Account")
}

func (Stripe) EventID(body []byte, _ *http.Request) string {
	var e struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &e)
	return e.ID
}

// parseStripeSignature pulls t and v1 from the Stripe-Signature header. Pure.
func parseStripeSignature(header string) (ts, v1 string) {
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}
	return ts, v1
}

// Verify checks the Stripe-Signature HMAC over "<t>.<body>".
func (Stripe) Verify(secret string, body []byte, r *http.Request) error {
	ts, v1 := parseStripeSignature(r.Header.Get(headerStripeSignature))
	if ts == "" || v1 == "" {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(v1), []byte(want)) {
		return ErrInvalidSignature
	}
	return nil
}

// stripeEvent is the subset of a Stripe event we consume. Sales are derived from
// charge.succeeded; SKU + quantity ride in charge metadata.
type stripeEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object struct {
			ID       string `json:"id"`
			Amount   int64  `json:"amount"` // minor units already
			Currency string `json:"currency"`
			Created  int64  `json:"created"`
			Metadata struct {
				SKU string `json:"sku"`
				Qty string `json:"qty"`
			} `json:"metadata"`
		} `json:"object"`
	} `json:"data"`
}

// ParseSales maps a charge.succeeded event to a single Sale. Other event types
// yield no sales (returns nil), so callers can ignore them safely.
func (Stripe) ParseSales(body []byte) ([]Sale, error) {
	var e stripeEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, fmt.Errorf("stripe: decode event: %w", err)
	}
	if e.Type != "charge.succeeded" {
		return nil, nil
	}
	obj := e.Data.Object
	qty := int32(1)
	if obj.Metadata.Qty != "" {
		if n, err := strconv.ParseInt(obj.Metadata.Qty, 10, 32); err == nil && n > 0 {
			qty = int32(n)
		}
	}
	return []Sale{{
		Source:      "stripe",
		OrderID:     obj.ID,
		SKU:         obj.Metadata.SKU,
		Qty:         qty,
		AmountCents: obj.Amount,
		Currency:    strings.ToUpper(obj.Currency),
		Channel:     "online",
		OccurredAt:  time.Unix(obj.Created, 0).UTC(),
	}}, nil
}
