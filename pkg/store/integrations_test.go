package store_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/store"
	"github.com/live-rack/pkg/store/internal/testhelper"
)

func TestIntegrationsRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping db integration test in short mode")
	}

	pool := testhelper.NewTestDB(t)
	ctx := context.Background()
	orgID := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO orgs (id, idp_org_id, name) VALUES ($1, $2, $3)`,
		orgID, "idp_test_"+orgID.String(), "Test Org")
	require.NoError(t, err)

	testhelper.SetOrgID(t, pool, orgID.String())
	q := store.New(pool)

	ig, err := q.UpsertIntegration(ctx, store.UpsertIntegrationParams{
		OrgID: orgID, Kind: "shopify", Status: "connected",
		ExternalID: "demo.myshopify.com", Secret: "shh", Config: []byte(`{}`),
	})
	require.NoError(t, err)
	assert.Equal(t, "connected", ig.Status)

	// Idempotency: routing function bypasses RLS.
	res, err := q.ResolveWebhookIntegration(ctx, store.ResolveWebhookIntegrationParams{
		Kind: "shopify", ExternalID: "demo.myshopify.com",
	})
	require.NoError(t, err)
	assert.Equal(t, orgID, res.OrgID)
	assert.Equal(t, "shh", res.Secret)

	first, err := q.InsertInboundWebhook(ctx, store.InsertInboundWebhookParams{
		OrgID: orgID, Provider: "shopify", EventID: "evt-1", Topic: "orders/create", Status: "received",
	})
	require.NoError(t, err)
	assert.Equal(t, "evt-1", first.EventID)

	// Duplicate delivery → ON CONFLICT DO NOTHING returns no row.
	_, err = q.InsertInboundWebhook(ctx, store.InsertInboundWebhookParams{
		OrgID: orgID, Provider: "shopify", EventID: "evt-1", Topic: "orders/create", Status: "received",
	})
	require.Error(t, err, "duplicate event id yields no returned row")

	rows, err := q.ListInboundWebhooks(ctx, store.ListInboundWebhooksParams{OrgID: orgID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 1)
}
