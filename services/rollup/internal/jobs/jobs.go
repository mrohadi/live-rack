// Package jobs defines the rollup job contract and a Runner that executes
// registered daily jobs against ClickHouse. The 5-minute zone-performance and
// 7x24 heatmap rollups are maintained by materialized views on insert (see
// pkg/chstore schema); jobs here cover aggregates that need cross-row joins.
package jobs

import (
	"context"
	"fmt"
	"time"
)

// Execer runs a ClickHouse statement that returns no rows. *chstore.Client
// satisfies it.
type Execer interface {
	Exec(ctx context.Context, sql string) error
}

// Job is a single daily rollup. SQL returns the statement that recomputes the
// job's target for the given day (jobs use ReplacingMergeTree targets, so
// re-running a day is idempotent).
type Job interface {
	Name() string
	SQL(day time.Time) string
}

// Runner executes a set of jobs in order.
type Runner struct {
	ch   Execer
	jobs []Job
}

// NewRunner builds a Runner over the given jobs.
func NewRunner(ch Execer, jobs ...Job) *Runner {
	return &Runner{ch: ch, jobs: jobs}
}

// Run executes every job for the given day, stopping at the first failure.
func (r *Runner) Run(ctx context.Context, day time.Time) error {
	for _, j := range r.jobs {
		if err := r.ch.Exec(ctx, j.SQL(day)); err != nil {
			return fmt.Errorf("rollup: job %s: %w", j.Name(), err)
		}
	}
	return nil
}

// DayString formats a day as ClickHouse Date literal (UTC). Pure.
func DayString(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}
