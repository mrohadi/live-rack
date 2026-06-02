// Package webhooks handles unauthenticated inbound POS webhooks (Shopify, Square).
// Each delivery is signature-verified, deduplicated, normalised to sales, and
// published to NATS for the ingest worker to persist.
package webhooks

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/integrations"
	"github.com/live-rack/pkg/store"
)

const maxWebhookBody = 1 << 20 // 1 MiB

// Store is the narrow store dependency the handler needs.
type Store interface {
	ResolveWebhookIntegration(ctx context.Context, arg store.ResolveWebhookIntegrationParams) (store.ResolveWebhookIntegrationRow, error)
	InsertInboundWebhook(ctx context.Context, arg store.InsertInboundWebhookParams) (store.WebhooksInbound, error)
	MarkWebhookStatus(ctx context.Context, arg store.MarkWebhookStatusParams) error
}

// Handler serves inbound webhook routes.
type Handler struct {
	stores   Store
	adapters map[string]integrations.Adapter
	pub      events.Publisher
}

// New builds a Handler from the given adapters.
func New(stores Store, pub events.Publisher, adapters ...integrations.Adapter) *Handler {
	m := make(map[string]integrations.Adapter, len(adapters))
	for _, a := range adapters {
		m[a.Kind()] = a
	}
	return &Handler{stores: stores, adapters: m, pub: pub}
}

// Register mounts webhook routes on e (no auth — verified by signature).
func (h *Handler) Register(e *echo.Echo) {
	e.POST("/webhooks/:provider", h.Receive)
}

// Receive verifies, deduplicates, and fans out one inbound webhook.
func (h *Handler) Receive(c echo.Context) error {
	provider := c.Param("provider")
	a, ok := h.adapters[provider]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "unknown provider")
	}

	req := c.Request()
	body, err := io.ReadAll(io.LimitReader(req.Body, maxWebhookBody))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "read body")
	}
	ctx := req.Context()

	handle := a.AccountHandle(body, req)
	if handle == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing account handle")
	}
	res, err := h.stores.ResolveWebhookIntegration(ctx, store.ResolveWebhookIntegrationParams{
		Kind: provider, ExternalID: handle,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unknown integration")
	}

	if err := a.Verify(res.Secret, body, req); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid signature")
	}

	eventID := a.EventID(body, req)
	if eventID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing event id")
	}

	rec, err := h.stores.InsertInboundWebhook(ctx, store.InsertInboundWebhookParams{
		OrgID: res.OrgID, Provider: provider, EventID: eventID,
		Topic: c.Request().Header.Get("X-Event-Topic"), Status: "received",
	})
	if err != nil {
		// ON CONFLICT DO NOTHING → no row → duplicate delivery; ack idempotently.
		return c.JSON(http.StatusOK, map[string]string{"status": "duplicate"})
	}

	sales, err := a.ParseSales(body)
	if err != nil {
		_ = h.stores.MarkWebhookStatus(ctx, store.MarkWebhookStatusParams{OrgID: res.OrgID, ID: rec.ID, Status: "rejected"})
		return echo.NewHTTPError(http.StatusBadRequest, "parse payload")
	}

	if err := h.publishSales(ctx, res.OrgID, provider, sales); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "publish sales")
	}

	if err := h.stores.MarkWebhookStatus(ctx, store.MarkWebhookStatusParams{OrgID: res.OrgID, ID: rec.ID, Status: "processed"}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "mark processed")
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok", "sales": len(sales)})
}

func (h *Handler) publishSales(ctx context.Context, orgID uuid.UUID, source string, sales []integrations.Sale) error {
	subject := events.POSSaleSubject(orgID)
	for _, s := range sales {
		if err := h.pub.Publish(ctx, subject, events.POSSale{
			OrgID:       orgID,
			Source:      source,
			OrderID:     s.OrderID,
			SKU:         s.SKU,
			Qty:         s.Qty,
			AmountCents: s.AmountCents,
			Currency:    s.Currency,
			Channel:     s.Channel,
			OccurredAt:  s.OccurredAt,
		}); err != nil {
			return err
		}
	}
	return nil
}
