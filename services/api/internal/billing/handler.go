// Package billing maps Stripe subscription webhooks to org plan changes. The
// Stripe price id selects the plan via a configured map; the org is taken from
// the subscription metadata (org_id). Signature verification reuses the Stripe
// adapter.
package billing

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/integrations"
)

// PlanUpdater persists an org's plan. *store.Queries satisfies it.
type PlanUpdater interface {
	UpdateOrgPlan(ctx context.Context, orgID uuid.UUID, plan string) error
}

// Handler serves the billing webhook.
type Handler struct {
	updater   PlanUpdater
	verifier  integrations.Stripe
	secret    string
	pricePlan map[string]domain.Plan
}

// New builds a Handler. priceToPlan maps Stripe price ids to plans.
func New(updater PlanUpdater, secret string, priceToPlan map[string]domain.Plan) *Handler {
	return &Handler{updater: updater, verifier: integrations.NewStripe(), secret: secret, pricePlan: priceToPlan}
}

// Register mounts the billing webhook (unauthenticated; Stripe-signed).
func (h *Handler) Register(e *echo.Echo) {
	e.POST("/webhooks/billing", h.Webhook)
}

// PlanForPrice resolves the plan for a Stripe price id, defaulting to free. Pure.
func (h *Handler) PlanForPrice(priceID string) domain.Plan {
	if p, ok := h.pricePlan[priceID]; ok {
		return p
	}
	return domain.PlanFree
}

type subscriptionEvent struct {
	Type string `json:"type"`
	Data struct {
		Object struct {
			Metadata struct {
				OrgID string `json:"org_id"`
			} `json:"metadata"`
			Items struct {
				Data []struct {
					Price struct {
						ID string `json:"id"`
					} `json:"price"`
				} `json:"data"`
			} `json:"items"`
		} `json:"object"`
	} `json:"data"`
}

// Webhook verifies the Stripe signature and applies the plan change.
func (h *Handler) Webhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "read body")
	}
	if err := h.verifier.Verify(h.secret, body, c.Request()); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid signature")
	}

	var e subscriptionEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "decode event")
	}
	switch e.Type {
	case "customer.subscription.created", "customer.subscription.updated":
		// handled below
	case "customer.subscription.deleted":
		// downgrade to free on cancellation
	default:
		return c.NoContent(http.StatusOK) // ignore unrelated events
	}

	orgID, err := uuid.Parse(e.Data.Object.Metadata.OrgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing org_id metadata")
	}

	plan := domain.PlanFree
	if e.Type != "customer.subscription.deleted" && len(e.Data.Object.Items.Data) > 0 {
		plan = h.PlanForPrice(e.Data.Object.Items.Data[0].Price.ID)
	}
	if err := h.updater.UpdateOrgPlan(c.Request().Context(), orgID, string(plan)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "update plan")
	}
	return c.JSON(http.StatusOK, map[string]string{"org_id": orgID.String(), "plan": string(plan)})
}
