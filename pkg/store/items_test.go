package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/store"
	"github.com/live-rack/pkg/store/internal/testhelper"
)

func TestItemsRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping db integration test in short mode")
	}

	pool := testhelper.NewTestDB(t)
	ctx := context.Background()

	orgID := uuid.New()
	storeID := uuid.New()
	zoneID := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO orgs (id, idp_org_id, name) VALUES ($1, $2, $3)`,
		orgID, "idp_test_"+orgID.String(), "Test Org")
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO stores (id, org_id, name) VALUES ($1, $2, $3)`,
		storeID, orgID, "Test Store")
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO zones (id, org_id, store_id, name) VALUES ($1, $2, $3, $4)`,
		zoneID, orgID, storeID, "Zone A")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, fmt.Sprintf("SET app.org_id = '%s'", orgID.String()))
	require.NoError(t, err)

	q := store.New(pool)

	t.Run("upsert item is idempotent on sku", func(t *testing.T) {
		a, err := q.UpsertItem(ctx, store.UpsertItemParams{
			OrgID: orgID, Sku: "SKU-1", Name: "Widget", Category: "frozen", Status: "active",
		})
		require.NoError(t, err)
		b, err := q.UpsertItem(ctx, store.UpsertItemParams{
			OrgID: orgID, Sku: "SKU-1", Name: "Widget v2", Category: "frozen", Status: "active",
		})
		require.NoError(t, err)
		assert.Equal(t, a.ID, b.ID)
		assert.Equal(t, "Widget v2", b.Name)
	})

	t.Run("adjust qty accumulates and floors at zero", func(t *testing.T) {
		l, err := q.AdjustItemLocationQty(ctx, store.AdjustItemLocationQtyParams{
			OrgID: orgID, StoreID: storeID, ZoneID: zoneID, Sku: "SKU-1", Qty: 5,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 5, l.Qty)

		l, err = q.AdjustItemLocationQty(ctx, store.AdjustItemLocationQtyParams{
			OrgID: orgID, StoreID: storeID, ZoneID: zoneID, Sku: "SKU-1", Qty: 3,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 8, l.Qty)

		l, err = q.AdjustItemLocationQty(ctx, store.AdjustItemLocationQtyParams{
			OrgID: orgID, StoreID: storeID, ZoneID: zoneID, Sku: "SKU-1", Qty: -100,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 0, l.Qty)
	})

	t.Run("list inventory joins item name", func(t *testing.T) {
		rows, err := q.ListInventoryByStore(ctx, store.ListInventoryByStoreParams{
			OrgID: orgID, StoreID: storeID,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "SKU-1", rows[0].Sku)
		assert.Equal(t, "Widget v2", rows[0].Name)
		assert.Equal(t, "frozen", rows[0].Category)
	})

	t.Run("velocity counts reflect rolling pick scans", func(t *testing.T) {
		now := time.Now()
		mk := func(action string, ts time.Time, valid bool) {
			_, err := q.CreateScanEvent(ctx, store.CreateScanEventParams{
				Ts: ts, OrgID: orgID, StoreID: storeID, ZoneID: zoneID,
				ScannerID: "s1", Sku: "SKU-1", Action: action, Valid: valid,
			})
			require.NoError(t, err)
		}
		mk("pick", now.Add(-1*time.Hour), true)     // in 7d
		mk("pick", now.Add(-2*24*time.Hour), true)  // in 7d
		mk("pick", now.Add(-20*24*time.Hour), true) // in 30d only
		mk("pick", now.Add(-2*time.Hour), false)    // invalid, ignored
		mk("place", now.Add(-1*time.Hour), true)    // not a pick, ignored

		rows, err := q.ListInventoryByStore(ctx, store.ListInventoryByStoreParams{
			OrgID: orgID, StoreID: storeID,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.EqualValues(t, 2, rows[0].Picks7d)
		assert.EqualValues(t, 3, rows[0].Picks30d)
	})
}
