package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const headerSquareSig = "X-Square-Hmacsha256-Signature"

// Square adapts Square order webhooks. Square signs base64(HMAC-SHA256(key,
// notificationURL+body)); the signing key is the integration secret.
type Square struct{}

// NewSquare builds a Square adapter.
func NewSquare() Square { return Square{} }

func (Square) Kind() string { return "square" }

// notificationURL reconstructs the URL Square signed. Square is always HTTPS.
func notificationURL(r *http.Request) string {
	return "https://" + r.Host + r.URL.RequestURI()
}

type squareEnvelope struct {
	MerchantID string `json:"merchant_id"`
	EventID    string `json:"event_id"`
	Type       string `json:"type"`
	CreatedAt  string `json:"created_at"`
	Data       struct {
		Object struct {
			Order struct {
				ID        string `json:"id"`
				LineItems []struct {
					CatalogObjectID string `json:"catalog_object_id"`
					Quantity        string `json:"quantity"`
					TotalMoney      struct {
						Amount   int64  `json:"amount"`
						Currency string `json:"currency"`
					} `json:"total_money"`
				} `json:"line_items"`
			} `json:"order"`
		} `json:"object"`
	} `json:"data"`
}

// AccountHandle returns merchant_id from the body to route the webhook.
func (Square) AccountHandle(body []byte, _ *http.Request) string {
	var e squareEnvelope
	if json.Unmarshal(body, &e) != nil {
		return ""
	}
	return e.MerchantID
}

// EventID returns event_id from the body for idempotency.
func (Square) EventID(body []byte, _ *http.Request) string {
	var e squareEnvelope
	if json.Unmarshal(body, &e) != nil {
		return ""
	}
	return e.EventID
}

// Verify checks the base64 HMAC-SHA256 of (notificationURL + body).
func (Square) Verify(secret string, body []byte, r *http.Request) error {
	got := r.Header.Get(headerSquareSig)
	if got == "" {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(notificationURL(r)))
	mac.Write(body)
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(got), []byte(want)) {
		return ErrInvalidSignature
	}
	return nil
}

// ParseSales normalises a Square order webhook into one Sale per line item.
func (Square) ParseSales(body []byte) ([]Sale, error) {
	var e squareEnvelope
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, fmt.Errorf("square: decode envelope: %w", err)
	}
	order := e.Data.Object.Order
	occurred := time.Now().UTC()
	if e.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, e.CreatedAt); err == nil {
			occurred = t.UTC()
		}
	}

	sales := make([]Sale, 0, len(order.LineItems))
	for _, li := range order.LineItems {
		if li.CatalogObjectID == "" {
			continue
		}
		qty, err := strconv.ParseInt(li.Quantity, 10, 32)
		if err != nil || qty <= 0 {
			continue
		}
		sales = append(sales, Sale{
			Source:      "square",
			OrderID:     order.ID,
			SKU:         li.CatalogObjectID,
			Qty:         int32(qty),
			AmountCents: li.TotalMoney.Amount,
			Currency:    li.TotalMoney.Currency,
			Channel:     "pos",
			OccurredAt:  occurred,
		})
	}
	return sales, nil
}
