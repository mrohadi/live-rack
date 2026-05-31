package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/domain"
)

func restorationDef() domain.PipelineDef {
	return domain.PipelineDef{
		Key:  "item-restoration",
		Name: "Item Restoration",
		Stages: []domain.StageDef{
			{Name: "Intake", SLA: 24 * time.Hour},
			{Name: "Triage", SLA: 48 * time.Hour},
			{Name: "Repair", SLA: 5 * 24 * time.Hour},
			{Name: "QA", SLA: 24 * time.Hour},
			{Name: "Restocked", SLA: 0},
		},
	}
}

func TestPipelineDef_Validate_OK(t *testing.T) {
	require.NoError(t, restorationDef().Validate())
}

func TestPipelineDef_Validate_Errors(t *testing.T) {
	cases := []struct {
		name string
		def  domain.PipelineDef
		want error
	}{
		{"no key", domain.PipelineDef{Name: "X", Stages: []domain.StageDef{{Name: "A"}}}, domain.ErrPipelineKeyRequired},
		{"no name", domain.PipelineDef{Key: "k", Stages: []domain.StageDef{{Name: "A"}}}, domain.ErrPipelineNameRequired},
		{"no stages", domain.PipelineDef{Key: "k", Name: "X"}, domain.ErrPipelineNoStages},
		{"blank stage", domain.PipelineDef{Key: "k", Name: "X", Stages: []domain.StageDef{{Name: " "}}}, domain.ErrStageNameRequired},
		{"negative sla", domain.PipelineDef{Key: "k", Name: "X", Stages: []domain.StageDef{{Name: "A", SLA: -1}}}, domain.ErrStageNegativeSLA},
		{"dup stage", domain.PipelineDef{Key: "k", Name: "X", Stages: []domain.StageDef{{Name: "A"}, {Name: "a"}}}, domain.ErrStageDuplicate},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.ErrorIs(t, c.def.Validate(), c.want)
		})
	}
}

func TestPipelineDef_StageIndex(t *testing.T) {
	d := restorationDef()
	assert.Equal(t, 0, d.StageIndex("Intake"))
	assert.Equal(t, 2, d.StageIndex("  repair "))
	assert.Equal(t, -1, d.StageIndex("Nope"))
}

func TestPipelineDef_Helpers(t *testing.T) {
	d := restorationDef()
	assert.Equal(t, 5, d.StageCount())
	assert.True(t, d.IsTerminalStage(4))
	assert.False(t, d.IsTerminalStage(0))
	assert.Equal(t, 5*24*time.Hour, d.SLAAt(2))
	assert.Equal(t, time.Duration(0), d.SLAAt(99))
}

func TestCardPriority_Valid(t *testing.T) {
	assert.True(t, domain.CardPriorityHigh.Valid())
	assert.False(t, domain.CardPriority("urgent").Valid())
}

func TestCanMutatePipeline(t *testing.T) {
	assert.True(t, domain.CanMutatePipeline(&domain.Principal{Role: domain.RoleAdmin}))
	assert.True(t, domain.CanMutatePipeline(&domain.Principal{Role: domain.RoleManager}))
	assert.False(t, domain.CanMutatePipeline(&domain.Principal{Role: domain.RoleStaff}))
	assert.False(t, domain.CanMutatePipeline(&domain.Principal{Role: domain.RoleReadonly}))
}
