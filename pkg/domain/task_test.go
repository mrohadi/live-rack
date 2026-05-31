package domain_test

import (
	"testing"

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
