package events_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/events"
)

func TestTaskSubject(t *testing.T) {
	org := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	assert.Equal(t, "lr.11111111-1111-1111-1111-111111111111.task.notified", events.TaskSubject(org))
}

func TestTaskSubject_RoundTripsOrgID(t *testing.T) {
	org := uuid.New()
	assert.Equal(t, org.String(), events.ExtractOrgID(events.TaskSubject(org)))
}
