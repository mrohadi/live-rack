package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/domain"
)

func TestWaveStatus_Valid(t *testing.T) {
	for _, s := range []domain.WaveStatus{
		domain.WaveOpen, domain.WavePicking, domain.WaveCompleted, domain.WaveCancelled,
	} {
		assert.True(t, s.Valid(), string(s))
	}
	assert.False(t, domain.WaveStatus("bogus").Valid())
}

func TestAllocatePick_FIFOFillsOrdersInOrder(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	demands := []domain.LineDemand{
		{LineID: a, Requested: 5},
		{LineID: b, Requested: 4},
		{LineID: c, Requested: 3},
	}
	// 7 picked: fills A (5), then B (2 of 4 → short), C gets 0 (short).
	got := domain.AllocatePick(7, demands)
	require.Len(t, got, 3)

	assert.Equal(t, a, got[0].LineID)
	assert.Equal(t, 5, got[0].Picked)
	assert.Equal(t, domain.PickLinePicked, got[0].Status)

	assert.Equal(t, 2, got[1].Picked)
	assert.Equal(t, domain.PickLineShort, got[1].Status)

	assert.Equal(t, 0, got[2].Picked)
	assert.Equal(t, domain.PickLineShort, got[2].Status)
}

func TestAllocatePick_FullSupplyAllPicked(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	got := domain.AllocatePick(9, []domain.LineDemand{
		{LineID: a, Requested: 5}, {LineID: b, Requested: 4},
	})
	assert.Equal(t, 5, got[0].Picked)
	assert.Equal(t, 4, got[1].Picked)
	assert.Equal(t, domain.PickLinePicked, got[0].Status)
	assert.Equal(t, domain.PickLinePicked, got[1].Status)
}

func TestAllocatePick_ZeroSupplyAllShort(t *testing.T) {
	a := uuid.New()
	got := domain.AllocatePick(0, []domain.LineDemand{{LineID: a, Requested: 3}})
	assert.Equal(t, 0, got[0].Picked)
	assert.Equal(t, domain.PickLineShort, got[0].Status)
}
