package integrations_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

const squareOrderJSON = `{
  "merchant_id": "MERCH1",
  "event_id": "evt-sq-1",
  "type": "order.created",
  "created_at": "2026-05-31T11:00:00Z",
  "data": {"object": {"order": {
    "id": "ord-1",
    "line_items": [
      {"catalog_object_id": "SKU1", "quantity": "2", "total_money": {"amount": 3998, "currency": "USD"}},
      {"catalog_object_id": "", "quantity": "1", "total_money": {"amount": 100, "currency": "USD"}}
    ]
  }}}
}`

func squareReq(sig string) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/webhooks/square", nil)
	r.Host = "api.live-rack.co"
	if sig != "" {
		r.Header.Set("X-Square-Hmacsha256-Signature", sig)
	}
	return r
}

func squareSig(secret string, r *http.Request, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("https://" + r.Host + r.URL.RequestURI()))
	mac.Write([]byte(body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestSquare_Verify_SignsURLPlusBody(t *testing.T) {
	a := integrations.NewSquare()
	body := []byte(squareOrderJSON)
	r := squareReq("")
	r.Header.Set("X-Square-Hmacsha256-Signature", squareSig("sqkey", r, squareOrderJSON))

	require.NoError(t, a.Verify("sqkey", body, r))
	assert.ErrorIs(t, a.Verify("wrong", body, r), integrations.ErrInvalidSignature)
	assert.ErrorIs(t, a.Verify("sqkey", body, squareReq("")), integrations.ErrInvalidSignature)
}

func TestSquare_Routing(t *testing.T) {
	a := integrations.NewSquare()
	body := []byte(squareOrderJSON)
	assert.Equal(t, "MERCH1", a.AccountHandle(body, squareReq("")))
	assert.Equal(t, "evt-sq-1", a.EventID(body, squareReq("")))
}

func TestSquare_ParseSales(t *testing.T) {
	a := integrations.NewSquare()
	sales, err := a.ParseSales([]byte(squareOrderJSON))
	require.NoError(t, err)
	require.Len(t, sales, 1, "blank catalog id skipped")
	assert.Equal(t, "square", sales[0].Source)
	assert.Equal(t, "ord-1", sales[0].OrderID)
	assert.Equal(t, "SKU1", sales[0].SKU)
	assert.Equal(t, int32(2), sales[0].Qty)
	assert.Equal(t, int64(3998), sales[0].AmountCents)
	assert.Equal(t, "pos", sales[0].Channel)
}
