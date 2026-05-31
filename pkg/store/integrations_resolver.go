package store

import (
	"context"

	"github.com/google/uuid"
)

// ResolveWebhookIntegrationRow is the org + signing secret for an inbound webhook.
type ResolveWebhookIntegrationRow struct {
	OrgID  uuid.UUID `json:"org_id"`
	Secret string    `json:"secret"`
}

// ResolveWebhookIntegrationParams routes a vendor handle to its integration.
type ResolveWebhookIntegrationParams struct {
	Kind       string `json:"kind"`
	ExternalID string `json:"external_id"`
}

// ResolveWebhookIntegration maps a vendor handle to its org + signing secret via
// the SECURITY DEFINER resolver, bypassing RLS for unauthenticated webhook
// routing. Hand-written because sqlc cannot analyse table-returning functions.
func (q *Queries) ResolveWebhookIntegration(ctx context.Context, arg ResolveWebhookIntegrationParams) (ResolveWebhookIntegrationRow, error) {
	row := q.db.QueryRow(ctx,
		`SELECT org_id, secret FROM resolve_webhook_integration($1, $2)`,
		arg.Kind, arg.ExternalID)
	var i ResolveWebhookIntegrationRow
	err := row.Scan(&i.OrgID, &i.Secret)
	return i, err
}
