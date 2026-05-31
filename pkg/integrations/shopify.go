package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

// Shopify HTTP headers.
const (
	headerShopifyHMAC    = "X-Shopify-Hmac-Sha256"
	headerShopifyEventID = "X-Shopify-Webhook-Id"
	headerShopifyShop    = "X-Shopify-Shop-Domain"
)

// Shopify adapts Shopify order webhooks. Signature is base64(HMAC-SHA256(body)).
type Shopify struct{}

// NewShopify builds a Shopify adapter.
func NewShopify() Shopify { return Shopify{} }

func (Shopify) Kind() string { return "shopify" }

// AccountHandle returns the shop domain header used to route the webhook.
func (Shopify) AccountHandle(_ []byte, r *http.Request) string {
	return r.Header.Get(headerShopifyShop)
}

func (Shopify) EventID(_ []byte, r *http.Request) string { return r.Header.Get(headerShopifyEventID) }

// Verify checks the X-Shopify-Hmac-Sha256 base64 digest against secret.
func (Shopify) Verify(secret string, body []byte, r *http.Request) error {
	got := r.Header.Get(headerShopifyHMAC)
	if got == "" {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(got), []byte(want)) {
		return ErrInvalidSignature
	}
	return nil
}

type shopifyOrder struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"created_at"`
	LineItems []struct {
		SKU      string `json:"sku"`
		Quantity int32  `json:"quantity"`
		Price    string `json:"price"` // per-unit, decimal string
	} `json:"line_items"`
}

// ParseSales normalises a Shopify order payload into one Sale per line item.
func (Shopify) ParseSales(body []byte) ([]Sale, error) {
	var o shopifyOrder
	if err := json.Unmarshal(body, &o); err != nil {
		return nil, fmt.Errorf("shopify: decode order: %w", err)
	}
	orderID := o.Name
	if orderID == "" {
		orderID = strconv.FormatInt(o.ID, 10)
	}
	occurred := time.Now().UTC()
	if o.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, o.CreatedAt); err == nil {
			occurred = t.UTC()
		}
	}

	sales := make([]Sale, 0, len(o.LineItems))
	for _, li := range o.LineItems {
		if li.SKU == "" || li.Quantity <= 0 {
			continue
		}
		unit, err := strconv.ParseFloat(li.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("shopify: parse price %q: %w", li.Price, err)
		}
		sales = append(sales, Sale{
			Source:      "shopify",
			OrderID:     orderID,
			SKU:         li.SKU,
			Qty:         li.Quantity,
			AmountCents: int64(math.Round(unit*100)) * int64(li.Quantity),
			Currency:    o.Currency,
			Channel:     "online",
			OccurredAt:  occurred,
		})
	}
	return sales, nil
}
