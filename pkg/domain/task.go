package domain

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus is a kanban column.
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusReview, TaskStatusDone:
		return true
	default:
		return false
	}
}

// TaskPriority orders board cards.
type TaskPriority string

const (
	TaskPriorityLow  TaskPriority = "low"
	TaskPriorityMed  TaskPriority = "med"
	TaskPriorityHigh TaskPriority = "high"
)

// Task is a unit of floor work, scoped to a store, optionally a zone.
type Task struct {
	ID         uuid.UUID    `json:"id"`
	OrgID      uuid.UUID    `json:"org_id"`
	StoreID    uuid.UUID    `json:"store_id"`
	ZoneID     *uuid.UUID   `json:"zone_id,omitempty"`
	Title      string       `json:"title"`
	Status     TaskStatus   `json:"status"`
	Priority   TaskPriority `json:"priority"`
	AssigneeID *uuid.UUID   `json:"assignee_id,omitempty"`
	DueAt      *time.Time   `json:"due_at,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// CanMutateTask reports whether the principal may create/update/delete tasks.
// Read-only and service principals are denied.
func CanMutateTask(p *Principal) bool {
	return p.HasRole(RoleAdmin, RoleManager, RoleStaff)
}
