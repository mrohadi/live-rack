package webhooks_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/integrations"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/webhooks"
)

type fakeStore struct {
	resolveOrg   uuid.UUID
	resolveSec   string
	resolveErr   error
	insertDup    bool // simulate ON CONFLICT DO NOTHING (no row)
	insertedRows int
	statuses     []string
}

func (f *fakeStore) ResolveWebhookIntegration(_ context.Context, _ store.ResolveWebhookIntegrationParams) (store.ResolveWebhookIntegrationRow, error) {
	if f.resolveErr != nil {
		return store.ResolveWebhookIntegrationRow{}, f.resolveErr
	}
	return store.ResolveWebhookIntegrationRow{OrgID: f.resolveOrg, Secret: f.resolveSec}, nil
}

func (f *fakeStore) InsertInboundWebhook(_ context.Context, arg store.InsertInboundWebhookParams) (store.WebhooksInbound, error) {
	if f.insertDup {
		return store.WebhooksInbound{}, errNoRows
	}
	f.insertedRows++
	return store.WebhooksInbound{ID: uuid.New(), OrgID: arg.OrgID, Provider: arg.Provider, EventID: arg.EventID}, nil
}

func (f *fakeStore) MarkWebhookStatus(_ context.Context, arg store.MarkWebhookStatusParams) error {
	f.statuses = append(f.statuses, arg.Status)
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }

const errNoRows = errString("no rows")

type fakePublisher struct {
	subjects []string
	payloads []any
}

func (p *fakePublisher) Publish(_ context.Context, subject string, v any) error {
	p.subjects = append(p.subjects, subject)
	p.payloads = append(p.payloads, v)
	return nil
}

const orderJSON = `{"id":1,"name":"#1001","currency":"USD","created_at":"2026-05-31T10:00:00Z","line_items":[{"sku":"LR-1240","quantity":2,"price":"19.99"}]}`

func sig(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func post(t *testing.T, h *webhooks.Handler, body string, headers map[string]string) (*httptest.ResponseRecorder, error) {
	t.Helper()
	e := echo.New()
	h.Register(e)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/webhooks/shopify", strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec, nil
}

func TestWebhook_HappyPath_PublishesSale(t *testing.T) {
	org := uuid.New()
	fs := &fakeStore{resolveOrg: org, resolveSec: "topsecret"}
	pub := &fakePublisher{}
	h := webhooks.New(fs, pub, integrations.NewShopify())

	rec, _ := post(t, h, orderJSON, map[string]string{
		"X-Shopify-Shop-Domain": "demo.myshopify.com",
		"X-Shopify-Hmac-Sha256": sig("topsecret", orderJSON),
		"X-Shopify-Webhook-Id":  "evt-1",
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, pub.payloads, 1)
	assert.Equal(t, events.POSSaleSubject(org), pub.subjects[0])
	sale := pub.payloads[0].(events.POSSale)
	assert.Equal(t, "LR-1240", sale.SKU)
	assert.Equal(t, int64(3998), sale.AmountCents)
	assert.Equal(t, []string{"processed"}, fs.statuses)
}

func TestWebhook_BadSignature_401(t *testing.T) {
	fs := &fakeStore{resolveOrg: uuid.New(), resolveSec: "topsecret"}
	pub := &fakePublisher{}
	h := webhooks.New(fs, pub, integrations.NewShopify())

	rec, _ := post(t, h, orderJSON, map[string]string{
		"X-Shopify-Shop-Domain": "demo.myshopify.com",
		"X-Shopify-Hmac-Sha256": sig("wrong", orderJSON),
		"X-Shopify-Webhook-Id":  "evt-1",
	})

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, pub.payloads)
}

func TestWebhook_Duplicate_AcksWithoutPublish(t *testing.T) {
	fs := &fakeStore{resolveOrg: uuid.New(), resolveSec: "topsecret", insertDup: true}
	pub := &fakePublisher{}
	h := webhooks.New(fs, pub, integrations.NewShopify())

	rec, _ := post(t, h, orderJSON, map[string]string{
		"X-Shopify-Shop-Domain": "demo.myshopify.com",
		"X-Shopify-Hmac-Sha256": sig("topsecret", orderJSON),
		"X-Shopify-Webhook-Id":  "evt-1",
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "duplicate")
	assert.Empty(t, pub.payloads)
}

func TestWebhook_UnknownProvider_404(t *testing.T) {
	fs := &fakeStore{}
	h := webhooks.New(fs, &fakePublisher{}, integrations.NewShopify())
	e := echo.New()
	h.Register(e)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/webhooks/stripe", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
