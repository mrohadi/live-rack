package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TaskNotifyKind labels why a task notification fired.
type TaskNotifyKind string

const (
	// TaskNotifyAssigned fires when a task gains an assignee.
	TaskNotifyAssigned TaskNotifyKind = "assigned"
	// TaskNotifyDeadline fires when an assigned task is due soon.
	TaskNotifyDeadline TaskNotifyKind = "deadline"
)

// TaskNotified is published when a task should notify its assignee.
type TaskNotified struct {
	OrgID      uuid.UUID      `json:"org_id"`
	StoreID    uuid.UUID      `json:"store_id"`
	TaskID     uuid.UUID      `json:"task_id"`
	AssigneeID uuid.UUID      `json:"assignee_id"`
	Kind       TaskNotifyKind `json:"kind"`
	Title      string         `json:"title"`
	DueAt      *time.Time     `json:"due_at,omitempty"`
	TS         time.Time      `json:"ts"`
}

const subjectTaskNotified = "lr.%s.task.notified"

// TaskSubject returns the per-org task.notified subject.
func TaskSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectTaskNotified, orgID)
}
