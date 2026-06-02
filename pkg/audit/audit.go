// Package audit writes append-only audit entries into the month-partitioned
// audit_log table. The partition-naming and DDL helpers plus entry validation
// are pure; the Postgres writer owns the I/O.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry is one audit record. Metadata is free-form JSON context.
type Entry struct {
	OrgID        uuid.UUID
	ActorUserID  uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
	TS           time.Time
}

// NewEntry validates and normalises an entry, defaulting TS to now. Pure.
func NewEntry(e Entry) (Entry, error) {
	if e.OrgID == uuid.Nil {
		return Entry{}, fmt.Errorf("audit: entry missing org_id")
	}
	if e.Action == "" || e.ResourceType == "" {
		return Entry{}, fmt.Errorf("audit: entry missing action/resource_type")
	}
	if e.TS.IsZero() {
		e.TS = time.Now()
	}
	e.TS = e.TS.UTC()
	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}
	return e, nil
}

// PartitionName returns the monthly partition table name for a timestamp. Pure.
func PartitionName(ts time.Time) string {
	return fmt.Sprintf("audit_log_%04d_%02d", ts.UTC().Year(), int(ts.UTC().Month()))
}

// PartitionDDL returns idempotent DDL creating the month partition for ts. Pure.
func PartitionDDL(ts time.Time) string {
	t := ts.UTC()
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_log FOR VALUES FROM ('%s') TO ('%s')",
		PartitionName(ts), start.Format("2006-01-02"), end.Format("2006-01-02"))
}

// Writer persists audit entries.
type Writer struct {
	pool *pgxpool.Pool
}

// NewWriter builds a Writer over a pgx pool.
func NewWriter(pool *pgxpool.Pool) *Writer {
	return &Writer{pool: pool}
}

// Write validates the entry, ensures its month partition exists, and inserts it.
func (w *Writer) Write(ctx context.Context, in Entry) error {
	e, err := NewEntry(in)
	if err != nil {
		return err
	}
	if _, err := w.pool.Exec(ctx, PartitionDDL(e.TS)); err != nil {
		return fmt.Errorf("audit: ensure partition: %w", err)
	}
	meta, err := json.Marshal(e.Metadata)
	if err != nil {
		return fmt.Errorf("audit: marshal metadata: %w", err)
	}
	var actor any
	if e.ActorUserID != uuid.Nil {
		actor = e.ActorUserID
	}
	if _, err := w.pool.Exec(ctx,
		`INSERT INTO audit_log (ts, org_id, actor_user_id, action, resource_type, resource_id, metadata)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.TS, e.OrgID, actor, e.Action, e.ResourceType, e.ResourceID, meta,
	); err != nil {
		return fmt.Errorf("audit: insert: %w", err)
	}
	return nil
}
