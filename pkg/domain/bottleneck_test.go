package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/domain"
)

func bnStages() []domain.StageDef {
	return []domain.StageDef{
		{Name: "Intake", SLA: time.Hour},
		{Name: "Repair", SLA: 2 * time.Hour},
		{Name: "Restocked", SLA: 0}, // terminal, never bottlenecks
	}
}

func TestDetectBottleneck_NoneAgeing(t *testing.T) {
	cards := []domain.CardAge{
		{StagePosition: 0, Age: 30 * time.Minute},
		{StagePosition: 1, Age: time.Hour},
	}
	assert.Nil(t, domain.DetectBottleneck(bnStages(), cards))
}

func TestDetectBottleneck_PicksMostBreaching(t *testing.T) {
	cards := []domain.CardAge{
		{StagePosition: 0, Age: 90 * time.Minute}, // breach intake
		{StagePosition: 1, Age: 3 * time.Hour},    // breach repair
		{StagePosition: 1, Age: 5 * time.Hour},    // breach repair
	}
	b := domain.DetectBottleneck(bnStages(), cards)
	require.NotNil(t, b)
	assert.Equal(t, 1, b.Position)
	assert.Equal(t, "Repair", b.Name)
	assert.Equal(t, 2, b.AgeingCount)
	assert.Equal(t, 5*time.Hour, b.OldestAge)
}

func TestDetectBottleneck_TieBreaksOnOldest(t *testing.T) {
	cards := []domain.CardAge{
		{StagePosition: 0, Age: 4 * time.Hour}, // intake, very old
		{StagePosition: 1, Age: 3 * time.Hour}, // repair
	}
	b := domain.DetectBottleneck(bnStages(), cards)
	require.NotNil(t, b)
	assert.Equal(t, 0, b.Position, "equal counts → oldest card wins")
}

func TestDetectBottleneck_TerminalStageIgnored(t *testing.T) {
	cards := []domain.CardAge{{StagePosition: 2, Age: 100 * time.Hour}}
	assert.Nil(t, domain.DetectBottleneck(bnStages(), cards))
}

func TestDetectBottleneck_OutOfRangeIgnored(t *testing.T) {
	cards := []domain.CardAge{{StagePosition: 9, Age: 100 * time.Hour}}
	assert.Nil(t, domain.DetectBottleneck(bnStages(), cards))
}
