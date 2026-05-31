package audit_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/audit"
)

func TestNewEntry(t *testing.T) {
	org := uuid.New()
	e, err := audit.NewEntry(audit.Entry{OrgID: org, Action: "task.create", ResourceType: "task"})
	require.NoError(t, err)
	assert.False(t, e.TS.IsZero())
	assert.NotNil(t, e.Metadata)

	_, err = audit.NewEntry(audit.Entry{Action: "x", ResourceType: "y"})
	assert.Error(t, err, "missing org")
	_, err = audit.NewEntry(audit.Entry{OrgID: org, ResourceType: "y"})
	assert.Error(t, err, "missing action")
}

func TestPartitionNameAndDDL(t *testing.T) {
	ts := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, "audit_log_2026_06", audit.PartitionName(ts))
	ddl := audit.PartitionDDL(ts)
	assert.Contains(t, ddl, "audit_log_2026_06 PARTITION OF audit_log")
	assert.Contains(t, ddl, "FROM ('2026-06-01') TO ('2026-07-01')")
}

// TestWriter_Integration writes an entry and reads it back. Skipped unless
// DATABASE_URL is set (migrations 0012 must be applied).
//
//	DATABASE_URL=postgres://postgres:postgres@localhost:5432/liverack?sslmode=disable \
//	go test ./pkg/audit/ -run Integration -v
func TestWriter_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping audit integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	org := uuid.New()
	w := audit.NewWriter(pool)
	require.NoError(t, w.Write(ctx, audit.Entry{
		OrgID: org, ActorUserID: uuid.New(), Action: "user.role.change",
		ResourceType: "user", ResourceID: "u-1", Metadata: map[string]any{"to": "admin"},
		TS: time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC),
	}))

	var action, meta string
	err = pool.QueryRow(ctx,
		"SELECT action, metadata::text FROM audit_log WHERE org_id=$1", org).Scan(&action, &meta)
	require.NoError(t, err)
	assert.Equal(t, "user.role.change", action)
	assert.Contains(t, meta, "admin")
}
