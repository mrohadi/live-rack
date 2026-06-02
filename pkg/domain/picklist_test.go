package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/domain"
)

func TestPickListStatus_Valid(t *testing.T) {
	for _, s := range []domain.PickListStatus{
		domain.PickListOpen, domain.PickListPicking,
		domain.PickListCompleted, domain.PickListCancelled,
	} {
		assert.True(t, s.Valid(), string(s))
	}
	assert.False(t, domain.PickListStatus("bogus").Valid())
}

func TestPickLineOutcome(t *testing.T) {
	assert.Equal(t, domain.PickLinePicked, domain.PickLineOutcome(5, 5))
	assert.Equal(t, domain.PickLinePicked, domain.PickLineOutcome(5, 6))
	assert.Equal(t, domain.PickLineShort, domain.PickLineOutcome(5, 3))
	assert.Equal(t, domain.PickLineShort, domain.PickLineOutcome(5, 0))
}

func TestPickShortfall(t *testing.T) {
	assert.Equal(t, 2, domain.PickShortfall(5, 3))
	assert.Equal(t, 0, domain.PickShortfall(5, 5))
	assert.Equal(t, 0, domain.PickShortfall(5, 9))
}
