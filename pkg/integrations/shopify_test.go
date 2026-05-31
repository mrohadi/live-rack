package integrations_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

func shopifySig(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

const shopifyOrderJSON = `{
  "id": 820982911946154508,
  "name": "#1001",
  "currency": "USD",
  "created_at": "2026-05-31T10:00:00Z",
  "line_items": [
    {"sku": "LR-1240", "quantity": 2, "price": "19.99"},
    {"sku": "LR-3318", "quantity": 1, "price": "5.00"},
    {"sku": "", "quantity": 3, "price": "1.00"}
  ]
}`

func TestShopify_Verify(t *testing.T) {
	a := integrations.NewShopify()
	body := []byte(shopifyOrderJSON)
	h := http.Header{}
	h.Set("X-Shopify-Hmac-Sha256", shopifySig("topsecret", body))

	require.NoError(t, a.Verify("topsecret", body, h))
	assert.ErrorIs(t, a.Verify("wrong", body, h), integrations.ErrInvalidSignature)

	empty := http.Header{}
	assert.ErrorIs(t, a.Verify("topsecret", body, empty), integrations.ErrInvalidSignature)
}

func TestShopify_HeaderHelpers(t *testing.T) {
	a := integrations.NewShopify()
	h := http.Header{}
	h.Set("X-Shopify-Webhook-Id", "evt-9")
	h.Set("X-Shopify-Shop-Domain", "demo.myshopify.com")
	assert.Equal(t, "evt-9", a.EventID(h))
	assert.Equal(t, "demo.myshopify.com", a.ShopDomain(h))
}

func TestShopify_ParseSales(t *testing.T) {
	a := integrations.NewShopify()
	sales, err := a.ParseSales([]byte(shopifyOrderJSON))
	require.NoError(t, err)
	require.Len(t, sales, 2, "blank-SKU line skipped")

	assert.Equal(t, "shopify", sales[0].Source)
	assert.Equal(t, "#1001", sales[0].OrderID)
	assert.Equal(t, "LR-1240", sales[0].SKU)
	assert.Equal(t, int32(2), sales[0].Qty)
	assert.Equal(t, int64(3998), sales[0].AmountCents) // 19.99 * 100 * 2
	assert.Equal(t, "USD", sales[0].Currency)
	assert.Equal(t, "online", sales[0].Channel)
}
