package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/pkg/store/internal/testhelper"
)

func TestZonesCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping db integration test in short mode")
	}

	pool := testhelper.NewTestDB(t)
	ctx := context.Background()

	// seed an org + store to satisfy FK constraints and RLS
	orgID := uuid.New()
	storeID := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO orgs (id, idp_org_id, name) VALUES ($1, $2, $3)`,
		orgID, "idp_test_"+orgID.String(), "Test Org",
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO stores (id, org_id, name) VALUES ($1, $2, $3)`,
		storeID, orgID, "Test Store",
	)
	require.NoError(t, err)

	// SET does not support parameterized values in Postgres
	_, err = pool.Exec(ctx, fmt.Sprintf("SET app.org_id = '%s'", orgID.String()))
	require.NoError(t, err)

	q := store.New(pool)

	t.Run("create zone", func(t *testing.T) {
		z, err := q.CreateZone(ctx, store.CreateZoneParams{
			OrgID:       orgID,
			StoreID:     storeID,
			Name:        "Zone A",
			Type:        store.ZoneTypeGeneral,
			X:           10,
			Y:           20,
			Width:       100,
			Height:      80,
			Color:       "#6366f1",
			Capacity:    50,
			Constraints: []byte(`{}`),
		})
		require.NoError(t, err)
		assert.Equal(t, "Zone A", z.Name)
		assert.Equal(t, store.ZoneTypeGeneral, z.Type)
		assert.Equal(t, orgID, z.OrgID)
		assert.Equal(t, storeID, z.StoreID)
	})

	t.Run("get zone", func(t *testing.T) {
		created, err := q.CreateZone(ctx, store.CreateZoneParams{
			OrgID: orgID, StoreID: storeID,
			Name: "Zone B", Type: store.ZoneTypeFrozen,
			X: 0, Y: 0, Width: 50, Height: 50,
			Color: "#0ea5e9", Capacity: 10,
			Constraints: []byte(`{}`),
		})
		require.NoError(t, err)

		got, err := q.GetZone(ctx, store.GetZoneParams{ID: created.ID, OrgID: orgID})
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "Zone B", got.Name)
	})

	t.Run("list zones by store", func(t *testing.T) {
		_, err := q.CreateZone(ctx, store.CreateZoneParams{
			OrgID: orgID, StoreID: storeID,
			Name: "Zone C", Type: store.ZoneTypeStaging,
			X: 0, Y: 0, Width: 40, Height: 40,
			Color: "#f59e0b", Capacity: 0,
			Constraints: []byte(`{}`),
		})
		require.NoError(t, err)

		zones, err := q.ListZonesByStore(ctx, store.ListZonesByStoreParams{
			StoreID: storeID, OrgID: orgID,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(zones), 1)
	})

	t.Run("update zone", func(t *testing.T) {
		created, err := q.CreateZone(ctx, store.CreateZoneParams{
			OrgID: orgID, StoreID: storeID,
			Name: "Zone D", Type: store.ZoneTypeGeneral,
			X: 0, Y: 0, Width: 60, Height: 60,
			Color: "#6366f1", Capacity: 20,
			Constraints: []byte(`{}`),
		})
		require.NoError(t, err)

		updated, err := q.UpdateZone(ctx, store.UpdateZoneParams{
			ID: created.ID, OrgID: orgID,
			Name: "Zone D Updated", Type: store.ZoneTypeReturns,
			X: 5, Y: 5, Width: 70, Height: 70,
			Color: "#ef4444", Capacity: 30,
			Constraints: []byte(`{"allowed_types":["returns"]}`),
		})
		require.NoError(t, err)
		assert.Equal(t, "Zone D Updated", updated.Name)
		assert.Equal(t, store.ZoneTypeReturns, updated.Type)
		assert.Equal(t, int32(30), updated.Capacity)
	})

	t.Run("delete zone", func(t *testing.T) {
		created, err := q.CreateZone(ctx, store.CreateZoneParams{
			OrgID: orgID, StoreID: storeID,
			Name: "Zone E", Type: store.ZoneTypeCheckout,
			X: 0, Y: 0, Width: 30, Height: 30,
			Color: "#10b981", Capacity: 5,
			Constraints: []byte(`{}`),
		})
		require.NoError(t, err)

		err = q.DeleteZone(ctx, store.DeleteZoneParams{ID: created.ID, OrgID: orgID})
		require.NoError(t, err)

		_, err = q.GetZone(ctx, store.GetZoneParams{ID: created.ID, OrgID: orgID})
		assert.Error(t, err, "expected not found error after delete")
	})

	t.Run("RLS isolation — other org cannot see zone", func(t *testing.T) {
		created, err := q.CreateZone(ctx, store.CreateZoneParams{
			OrgID: orgID, StoreID: storeID,
			Name: "RLS Zone", Type: store.ZoneTypeGeneral,
			X: 0, Y: 0, Width: 100, Height: 100,
			Color: "#6366f1", Capacity: 0,
			Constraints: []byte(`{}`),
		})
		require.NoError(t, err)

		// Acquire a dedicated connection, switch to app role (subject to RLS),
		// then set a different org_id — zone must not be visible.
		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)
		defer func() {
			// reset session state before returning connection to pool
			_, _ = conn.Exec(ctx, "RESET ROLE")
			_, _ = conn.Exec(ctx, fmt.Sprintf("SET app.org_id = '%s'", orgID.String()))
			conn.Release()
		}()

		otherOrgID := uuid.New()
		_, err = conn.Exec(ctx, "SET ROLE liverack_app")
		require.NoError(t, err)
		_, err = conn.Exec(ctx, fmt.Sprintf("SET app.org_id = '%s'", otherOrgID.String()))
		require.NoError(t, err)

		appQ := store.New(conn)
		_, err = appQ.GetZone(ctx, store.GetZoneParams{ID: created.ID, OrgID: orgID})
		assert.Error(t, err, "other org must not see this zone")
	})

	t.Run("count zones by store", func(t *testing.T) {
		emptyStoreID := uuid.New()
		_, err := pool.Exec(ctx,
			`INSERT INTO stores (id, org_id, name) VALUES ($1, $2, $3)`,
			emptyStoreID, orgID, "Empty Store",
		)
		require.NoError(t, err)

		count, err := q.CountZonesByStore(ctx, store.CountZonesByStoreParams{
			StoreID: emptyStoreID, OrgID: orgID,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("constraints round-trip through JSONB", func(t *testing.T) {
		max := 25
		original := domain.ZoneConstraints{
			AllowedCategories: []string{"frozen", "dairy"},
			DeniedCategories:  []string{"hazmat"},
			MaxUnitsPerSKU:    &max,
			RequireDualScan:   true,
		}
		raw, err := domain.MarshalConstraints(original)
		require.NoError(t, err)

		created, err := q.CreateZone(ctx, store.CreateZoneParams{
			OrgID: orgID, StoreID: storeID,
			Name: "Constraint Zone", Type: store.ZoneTypeFrozen,
			X: 0, Y: 0, Width: 80, Height: 80,
			Color: "#0ea5e9", Capacity: 200,
			Constraints: raw,
		})
		require.NoError(t, err)

		got, err := q.GetZone(ctx, store.GetZoneParams{ID: created.ID, OrgID: orgID})
		require.NoError(t, err)

		parsed, err := domain.UnmarshalConstraints(got.Constraints)
		require.NoError(t, err)
		assert.Equal(t, original.AllowedCategories, parsed.AllowedCategories)
		assert.Equal(t, original.DeniedCategories, parsed.DeniedCategories)
		require.NotNil(t, parsed.MaxUnitsPerSKU)
		assert.Equal(t, max, *parsed.MaxUnitsPerSKU)
		assert.True(t, parsed.RequireDualScan)

		// Domain rule using stored bytes — proves the wire→DB→domain loop is intact.
		err = got.AsDomainZone().CanAcceptItem("hazmat", 0, 1)
		assert.ErrorIs(t, err, domain.ErrCategoryDenied)
	})
}
