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

func TestPipelinesRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping db integration test in short mode")
	}

	pool := testhelper.NewTestDB(t)
	ctx := context.Background()

	orgID := uuid.New()
	storeID := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO orgs (id, idp_org_id, name) VALUES ($1, $2, $3)`,
		orgID, "idp_test_"+orgID.String(), "Test Org")
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO stores (id, org_id, name) VALUES ($1, $2, $3)`,
		storeID, orgID, "Test Store")
	require.NoError(t, err)

	testhelper.SetOrgID(t, pool, orgID.String())

	q := store.New(pool)

	pipe, err := q.CreatePipeline(ctx, store.CreatePipelineParams{
		OrgID: orgID, StoreID: storeID, Key: "item-restoration", Name: "Item Restoration",
	})
	require.NoError(t, err)
	assert.Equal(t, "item-restoration", pipe.Key)

	for i, name := range []string{"Intake", "Triage", "Repair"} {
		_, err := q.CreateStage(ctx, store.CreateStageParams{
			OrgID: orgID, PipelineID: pipe.ID, Position: int32(i), Name: name,
			SlaSeconds: int64((i + 1) * 3600),
		})
		require.NoError(t, err)
	}
	stages, err := q.ListStagesByPipeline(ctx, store.ListStagesByPipelineParams{OrgID: orgID, PipelineID: pipe.ID})
	require.NoError(t, err)
	require.Len(t, stages, 3)
	assert.Equal(t, "Intake", stages[0].Name)

	card, err := q.CreateCard(ctx, store.CreateCardParams{
		OrgID: orgID, PipelineID: pipe.ID, StagePosition: 0,
		Title: "Espresso Machine · scratched casing", Sku: "LR-1240", Priority: "low",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), card.StagePosition)
	entered0 := card.EnteredStageAt

	moved, err := q.MoveCard(ctx, store.MoveCardParams{OrgID: orgID, ID: card.ID, StagePosition: 2})
	require.NoError(t, err)
	assert.Equal(t, int32(2), moved.StagePosition)
	assert.True(t, moved.EnteredStageAt.After(entered0) || moved.EnteredStageAt.Equal(entered0))

	cards, err := q.ListCardsByPipeline(ctx, store.ListCardsByPipelineParams{OrgID: orgID, PipelineID: pipe.ID})
	require.NoError(t, err)
	require.Len(t, cards, 1)
	assert.Equal(t, int32(2), cards[0].StagePosition)

	pipes, err := q.ListPipelinesByStore(ctx, store.ListPipelinesByStoreParams{OrgID: orgID, StoreID: storeID})
	require.NoError(t, err)
	require.Len(t, pipes, 1)
}
