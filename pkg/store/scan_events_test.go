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

func TestScanEventsRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping db integration test in short mode")
	}

	pool := testhelper.NewTestDB(t)
	ctx := context.Background()

	// seed an org + store to satisfy FK constraints and RLS
	orgID := uuid.New()
	storeID := uuid.New()
	zoneID := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO orgs (id, clerk_org_id, name) VALUES ($1, $2, $3)`,
		orgID, "clerk_test_"+orgID.String(), "Test Org")
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO stores (id, org_id, name) VALUES ($1, $2, $3)`,
		storeID, orgID, "Test Store")
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO zones (id, org_id, store_id, name) VALUES ($1, $2, $3, $4)`,
		zoneID, orgID, storeID, "Zone A")
	require.NoError(t, err)

	// SET does not support parameterized values in Postgres
	_, err = pool.Exec(ctx, fmt.Sprintf("SET app.org_id = '%s'", orgID.String()))
	require.NoError(t, err)

	q := store.New(pool)
	earlier := time.Now().Add(-time.Minute).UTC()

	t.Run("record scan", func(t *testing.T) {
		ev, err := q.CreateScanEvent(ctx, store.CreateScanEventParams{
			Ts:        earlier,
			OrgID:     orgID,
			StoreID:   storeID,
			ZoneID:    zoneID,
			ScannerID: "scanner-01",
			Sku:       "SKU-123",
			Action:    "place",
			Valid:     true,
		})
		require.NoError(t, err)
		assert.Equal(t, "SKU-123", ev.Sku)
		assert.True(t, ev.Valid)
	})

	t.Run("last scan for sku in zone", func(t *testing.T) {
		recent := time.Now().UTC()
		_, err := q.CreateScanEvent(ctx, store.CreateScanEventParams{
			Ts:        recent,
			OrgID:     orgID,
			StoreID:   storeID,
			ZoneID:    zoneID,
			ScannerID: "scanner-01",
			Sku:       "SKU-123",
			Action:    "place",
			Valid:     true,
		})
		require.NoError(t, err)

		last, err := q.GetLastScanForSKU(ctx, store.GetLastScanForSKUParams{
			OrgID:  orgID,
			ZoneID: zoneID,
			Sku:    "SKU-123",
		})
		require.NoError(t, err)
		assert.WithinDuration(t, recent, last.Ts, time.Second)
	})

	t.Run("list by zone newest first", func(t *testing.T) {
		rows, err := q.ListScanEventsByZone(ctx, store.ListScanEventsByZoneParams{
			OrgID:  orgID,
			ZoneID: zoneID,
			Limit:  10,
		})
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.True(t, !rows[0].Ts.Before(rows[1].Ts))
	})
}
