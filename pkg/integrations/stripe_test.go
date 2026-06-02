package integrations_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

func stripeSig(secret, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func stripeReq(sig string) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	r.Header.Set("Stripe-Signature", sig)
	r.Header.Set("Stripe-Account", "acct_123")
	return r
}

const chargeBody = `{"id":"evt_1","type":"charge.succeeded","data":{"object":{
	"id":"ch_1","amount":3998,"currency":"usd","created":1717200000,
	"metadata":{"sku":"LR-1240","qty":"2"}}}}`

func TestStripe_VerifyAndParse(t *testing.T) {
	body := []byte(chargeBody)
	r := stripeReq(stripeSig("whsec", "1717200000", body))

	s := integrations.NewStripe()
	require.NoError(t, s.Verify("whsec", body, r))
	assert.Equal(t, "acct_123", s.AccountHandle(body, r))
	assert.Equal(t, "evt_1", s.EventID(body, r))

	sales, err := s.ParseSales(body)
	require.NoError(t, err)
	require.Len(t, sales, 1)
	assert.Equal(t, "stripe", sales[0].Source)
	assert.Equal(t, "LR-1240", sales[0].SKU)
	assert.Equal(t, int32(2), sales[0].Qty)
	assert.Equal(t, int64(3998), sales[0].AmountCents)
	assert.Equal(t, "USD", sales[0].Currency)
}

func TestStripe_VerifyRejectsBadSig(t *testing.T) {
	body := []byte(chargeBody)
	r := stripeReq(stripeSig("wrong", "1717200000", body))
	assert.ErrorIs(t, integrations.NewStripe().Verify("whsec", body, r), integrations.ErrInvalidSignature)
}

func TestStripe_VerifyRejectsMissing(t *testing.T) {
	r := stripeReq("garbage")
	assert.ErrorIs(t, integrations.NewStripe().Verify("whsec", []byte(chargeBody), r), integrations.ErrInvalidSignature)
}

func TestStripe_IgnoresOtherEvents(t *testing.T) {
	sales, err := integrations.NewStripe().ParseSales([]byte(`{"id":"evt_2","type":"customer.created","data":{"object":{}}}`))
	require.NoError(t, err)
	assert.Empty(t, sales)
}

func TestStripe_DefaultQty(t *testing.T) {
	body := []byte(`{"type":"charge.succeeded","data":{"object":{"id":"ch_2","amount":500,"currency":"eur","metadata":{"sku":"X"}}}}`)
	sales, err := integrations.NewStripe().ParseSales(body)
	require.NoError(t, err)
	require.Len(t, sales, 1)
	assert.Equal(t, int32(1), sales[0].Qty)
	assert.Equal(t, "EUR", sales[0].Currency)
}
