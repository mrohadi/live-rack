// Package integrations serves the integrations list and inbound webhook event log.
package integrations

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/store"
)

const (
	defaultLogLimit = 50
	maxLogLimit     = 200
)

// Store is the narrow store dependency the handler needs.
type Store interface {
	ListIntegrations(ctx context.Context, orgID uuid.UUID) ([]store.ListIntegrationsRow, error)
	ListInboundWebhooks(ctx context.Context, arg store.ListInboundWebhooksParams) ([]store.WebhooksInbound, error)
}

// Handler serves integrations + webhook log endpoints.
type Handler struct {
	q Store
}

// New creates a Handler.
func New(q Store) *Handler {
	return &Handler{q: q}
}

// Register mounts integrations routes on the authenticated API group.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/integrations", h.List)
	g.GET("/integrations/webhooks", h.WebhookLog)
}

// IntegrationRow is one connected integration.
type IntegrationRow struct {
	ID         uuid.UUID `json:"id"`
	Kind       string    `json:"kind"`
	Status     string    `json:"status"`
	ExternalID string    `json:"external_id"`
}

// List godoc
//
//	@Summary	List the org's integrations
//	@Tags		integrations
//	@Produce	json
//	@Success	200	{array}	IntegrationRow
//	@Router		/integrations [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	rows, err := h.q.ListIntegrations(c.Request().Context(), p.OrgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list integrations")
	}
	out := make([]IntegrationRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, IntegrationRow{ID: r.ID, Kind: r.Kind, Status: r.Status, ExternalID: r.ExternalID})
	}
	return c.JSON(http.StatusOK, out)
}

// WebhookRow is one inbound webhook delivery.
type WebhookRow struct {
	ID         uuid.UUID `json:"id"`
	Provider   string    `json:"provider"`
	EventID    string    `json:"event_id"`
	Topic      string    `json:"topic"`
	Status     string    `json:"status"`
	ReceivedAt string    `json:"received_at"`
}

// WebhookLog godoc
//
//	@Summary	List recent inbound webhook deliveries
//	@Tags		integrations
//	@Produce	json
//	@Param		limit	query		int	false	"Max rows (default 50, max 200)"
//	@Success	200		{array}		WebhookRow
//	@Router		/integrations/webhooks [get]
func (h *Handler) WebhookLog(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	limit := defaultLogLimit
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLogLimit {
		limit = maxLogLimit
	}

	rows, err := h.q.ListInboundWebhooks(c.Request().Context(), store.ListInboundWebhooksParams{
		OrgID: p.OrgID, Limit: int32(limit),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list webhooks")
	}
	out := make([]WebhookRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, WebhookRow{
			ID: r.ID, Provider: r.Provider, EventID: r.EventID, Topic: r.Topic,
			Status: r.Status, ReceivedAt: r.ReceivedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return c.JSON(http.StatusOK, out)
}
