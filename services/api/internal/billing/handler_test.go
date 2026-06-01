package billing_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/services/api/internal/billing"
)

type fakeUpdater struct {
	gotOrg  uuid.UUID
	gotPlan string
	calls   int
}

func (f *fakeUpdater) UpdateOrgPlan(_ context.Context, orgID uuid.UUID, plan string) error {
	f.gotOrg = orgID
	f.gotPlan = plan
	f.calls++
	return nil
}

const secret = "whsec_test"

func stripeSig(ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func post(t *testing.T, h *billing.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	h.Register(e)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/webhooks/billing", strings.NewReader(body))
	req.Header.Set("Stripe-Signature", stripeSig("1717200000", []byte(body)))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func newHandler(u *fakeUpdater) *billing.Handler {
	return billing.New(u, secret, map[string]domain.Plan{
		"price_growth": domain.PlanGrowth,
		"price_ent":    domain.PlanEnterprise,
	})
}

func subEvent(org, typ, price string) string {
	return `{"type":"` + typ + `","data":{"object":{"metadata":{"org_id":"` + org +
		`"},"items":{"data":[{"price":{"id":"` + price + `"}}]}}}}`
}

func TestWebhook_UpgradesPlan(t *testing.T) {
	org := uuid.New()
	u := &fakeUpdater{}
	rec := post(t, newHandler(u), subEvent(org.String(), "customer.subscription.updated", "price_growth"))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, org, u.gotOrg)
	assert.Equal(t, "growth", u.gotPlan)
}

func TestWebhook_CancellationDowngradesToFree(t *testing.T) {
	org := uuid.New()
	u := &fakeUpdater{}
	rec := post(t, newHandler(u), subEvent(org.String(), "customer.subscription.deleted", "price_growth"))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "free", u.gotPlan)
}

func TestWebhook_UnknownPriceDefaultsFree(t *testing.T) {
	org := uuid.New()
	u := &fakeUpdater{}
	post(t, newHandler(u), subEvent(org.String(), "customer.subscription.created", "price_???"))
	assert.Equal(t, "free", u.gotPlan)
}

func TestWebhook_IgnoresUnrelatedEvents(t *testing.T) {
	u := &fakeUpdater{}
	rec := post(t, newHandler(u), `{"type":"invoice.paid","data":{"object":{}}}`)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 0, u.calls)
}

func TestWebhook_RejectsBadSignature(t *testing.T) {
	u := &fakeUpdater{}
	e := echo.New()
	newHandler(u).Register(e)
	body := subEvent(uuid.New().String(), "customer.subscription.updated", "price_growth")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/webhooks/billing", strings.NewReader(body))
	req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
