package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/live-rack/pkg/domain"
)

func TestCanMutateTask(t *testing.T) {
	cases := []struct {
		role domain.RoleName
		want bool
	}{
		{domain.RoleAdmin, true},
		{domain.RoleManager, true},
		{domain.RoleStaff, true},
		{domain.RoleReadonly, false},
		{domain.RoleService, false},
	}
	for _, c := range cases {
		p := &domain.Principal{OrgID: uuid.New(), Role: c.role}
		if got := domain.CanMutateTask(p); got != c.want {
			t.Errorf("role %s: got %v want %v", c.role, got, c.want)
		}
	}
}

func TestTaskStatusValid(t *testing.T) {
	if !domain.TaskStatusTodo.Valid() {
		t.Error("todo should be valid")
	}
	if domain.TaskStatus("bogus").Valid() {
		t.Error("bogus should be invalid")
	}
}

func TestTaskDueSoon(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	within := 24 * time.Hour
	in6h := now.Add(6 * time.Hour)
	in2d := now.Add(48 * time.Hour)
	past := now.Add(-time.Hour)

	cases := []struct {
		name string
		due  *time.Time
		want bool
	}{
		{"no due date", nil, false},
		{"due in 6h", &in6h, true},
		{"due in 2d", &in2d, false},
		{"already past", &past, false},
	}
	for _, c := range cases {
		got := domain.Task{DueAt: c.due}.DueSoon(now, within)
		if got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
