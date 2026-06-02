package integrations_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	integrationsapi "github.com/live-rack/services/api/internal/integrations"
)

type fakeStore struct {
	integrations []store.ListIntegrationsRow
	webhooks     []store.WebhooksInbound
	gotLimit     int32
}

func (f *fakeStore) ListIntegrations(_ context.Context, _ uuid.UUID) ([]store.ListIntegrationsRow, error) {
	return f.integrations, nil
}
func (f *fakeStore) ListInboundWebhooks(_ context.Context, arg store.ListInboundWebhooksParams) ([]store.WebhooksInbound, error) {
	f.gotLimit = arg.Limit
	return f.webhooks, nil
}

func serve(t *testing.T, h *integrationsapi.Handler, target string, p *domain.Principal) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	h.Register(e.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestIntegrations_List(t *testing.T) {
	fs := &fakeStore{integrations: []store.ListIntegrationsRow{
		{ID: uuid.New(), Kind: "shopify", Status: "connected", ExternalID: "demo.myshopify.com"},
	}}
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleAdmin}
	rec := serve(t, integrationsapi.New(fs), "/api/v1/integrations", p)

	require.Equal(t, http.StatusOK, rec.Code)
	var out []integrationsapi.IntegrationRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "shopify", out[0].Kind)
	assert.Equal(t, "connected", out[0].Status)
}

func TestIntegrations_WebhookLog(t *testing.T) {
	fs := &fakeStore{webhooks: []store.WebhooksInbound{
		{ID: uuid.New(), Provider: "shopify", EventID: "evt-1", Topic: "orders/create", Status: "processed", ReceivedAt: time.Now().UTC()},
	}}
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleManager}
	rec := serve(t, integrationsapi.New(fs), "/api/v1/integrations/webhooks?limit=10", p)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int32(10), fs.gotLimit)
	var out []integrationsapi.WebhookRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "processed", out[0].Status)
	assert.Equal(t, "orders/create", out[0].Topic)
}

func TestIntegrations_WebhookLog_LimitCapped(t *testing.T) {
	fs := &fakeStore{}
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleAdmin}
	serve(t, integrationsapi.New(fs), "/api/v1/integrations/webhooks?limit=9999", p)
	assert.Equal(t, int32(200), fs.gotLimit)
}
