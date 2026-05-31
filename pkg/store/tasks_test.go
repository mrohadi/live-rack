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

func TestTasksRepo(t *testing.T) {
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

	_, err = pool.Exec(ctx, "SET app.org_id = '"+orgID.String()+"'")
	require.NoError(t, err)

	q := store.New(pool)

	created, err := q.CreateTask(ctx, store.CreateTaskParams{
		OrgID: orgID, StoreID: storeID, Title: "Restock A1",
		Status: "todo", Priority: "high",
	})
	require.NoError(t, err)
	assert.Equal(t, "todo", created.Status)

	moved, err := q.UpdateTaskStatus(ctx, store.UpdateTaskStatusParams{
		OrgID: orgID, ID: created.ID, Status: "in_progress",
	})
	require.NoError(t, err)
	assert.Equal(t, "in_progress", moved.Status)

	rows, err := q.ListTasksByStore(ctx, store.ListTasksByStoreParams{
		OrgID: orgID, StoreID: storeID,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)

	require.NoError(t, q.DeleteTask(ctx, store.DeleteTaskParams{OrgID: orgID, ID: created.ID}))
}
