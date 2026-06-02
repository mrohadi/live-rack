package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/store"
	"github.com/live-rack/pkg/store/internal/testhelper"
)

func TestSalesRepo(t *testing.T) {
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

	now := time.Now().UTC()
	rows := []store.CreateSaleEventParams{
		{Ts: now, OrgID: orgID, Source: "shopify", OrderID: "#1", Sku: "A", Qty: 2, AmountCents: 4000, Currency: "USD", Channel: "online"},
		{Ts: now, OrgID: orgID, Source: "shopify", OrderID: "#1", Sku: "B", Qty: 1, AmountCents: 1000, Currency: "USD", Channel: "online"},
		{Ts: now.Add(-48 * time.Hour), OrgID: orgID, Source: "square", OrderID: "#2", Sku: "A", Qty: 1, AmountCents: 2000, Currency: "USD", Channel: "pos"},
	}
	for _, r := range rows {
		_, err := q.CreateSaleEvent(ctx, r)
		require.NoError(t, err)
	}

	// Summary for today only (last 24h) → first two rows.
	sum, err := q.SalesSummary(ctx, store.SalesSummaryParams{OrgID: orgID, Ts: now.Add(-24 * time.Hour)})
	require.NoError(t, err)
	assert.Equal(t, int64(5000), sum.RevenueCents)
	assert.Equal(t, int64(3), sum.Units)
	assert.Equal(t, int64(1), sum.Orders) // both lines share order #1

	byDay, err := q.SalesByDay(ctx, store.SalesByDayParams{OrgID: orgID, Ts: now.Add(-7 * 24 * time.Hour)})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(byDay), 2)
}
