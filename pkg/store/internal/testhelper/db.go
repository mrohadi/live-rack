package testhelper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName = "liverack_test"
	dbUser = "liverack_test"
	dbPass = "liverack_test"
)

// NewTestDB spins up a Postgres container, runs all goose migrations, and
// returns a pool scoped to an isolated org_id via app.org_id session variable.
func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("timescale/timescaledb:latest-pg16"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = ctr.Terminate(ctx) })

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	runMigrations(t, dsn)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open pgxpool: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// SetOrgID sets app.org_id on the connection so RLS policies resolve.
func SetOrgID(t *testing.T, pool *pgxpool.Pool, orgID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		fmt.Sprintf("SET app.org_id = '%s'", orgID))
	if err != nil {
		t.Fatalf("set app.org_id: %v", err)
	}
}

func runMigrations(t *testing.T, dsn string) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine source file path")
	}
	// internal/testhelper/db.go → ../../../../migrations
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations")
	migrationsDir = filepath.Clean(migrationsDir)

	if _, err := os.Stat(migrationsDir); err != nil {
		t.Fatalf("migrations dir not found at %s: %v", migrationsDir, err)
	}

	// goose via exec to avoid importing goose into store pkg
	t.Logf("running migrations from %s", migrationsDir)
	if err := runGoose(dsn, migrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
}
