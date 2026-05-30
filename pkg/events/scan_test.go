package events_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/events"
)

func TestScanSubject(t *testing.T) {
	org := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	got := events.ScanSubject(org)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111.scan.recorded", got)
}

func TestExtractOrgID(t *testing.T) {
	got := events.ExtractOrgID("11111111-1111-1111-1111-111111111111.scan.recorded")
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", got)
}
