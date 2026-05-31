package jobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/live-rack/services/rollup/internal/jobs"
)

type fakeJob struct {
	name string
	sql  string
}

func (f fakeJob) Name() string           { return f.name }
func (f fakeJob) SQL(_ time.Time) string { return f.sql }

type recExecer struct {
	got    []string
	failOn string
}

func (r *recExecer) Exec(_ context.Context, sql string) error {
	r.got = append(r.got, sql)
	if r.failOn != "" && sql == r.failOn {
		return errors.New("boom")
	}
	return nil
}

func TestRunner_RunsAllInOrder(t *testing.T) {
	ex := &recExecer{}
	r := jobs.NewRunner(ex, fakeJob{"a", "SQL_A"}, fakeJob{"b", "SQL_B"})
	if err := r.Run(context.Background(), time.Now()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(ex.got) != 2 || ex.got[0] != "SQL_A" || ex.got[1] != "SQL_B" {
		t.Errorf("got %#v, want [SQL_A SQL_B]", ex.got)
	}
}

func TestRunner_StopsOnError(t *testing.T) {
	ex := &recExecer{failOn: "SQL_A"}
	r := jobs.NewRunner(ex, fakeJob{"a", "SQL_A"}, fakeJob{"b", "SQL_B"})
	err := r.Run(context.Background(), time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
	if len(ex.got) != 1 {
		t.Errorf("expected to stop after first job, ran %d", len(ex.got))
	}
}

func TestDayString(t *testing.T) {
	d := time.Date(2026, 6, 1, 23, 59, 0, 0, time.UTC)
	if got := jobs.DayString(d); got != "2026-06-01" {
		t.Errorf("DayString = %q, want 2026-06-01", got)
	}
}
