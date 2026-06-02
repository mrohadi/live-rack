package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/domain"
)

func TestRestorationTemplate_IsValid(t *testing.T) {
	d := domain.RestorationTemplate()
	require.NoError(t, d.Validate())
	assert.Equal(t, domain.TemplateItemRestoration, d.Key)
	assert.Equal(t, 5, d.StageCount())
	assert.Equal(t, 0, d.StageIndex("Intake"))
	assert.Equal(t, 4, d.StageIndex("Restocked"))
	assert.Equal(t, 5*24*time.Hour, d.SLAAt(2)) // Repair
	assert.True(t, d.IsTerminalStage(4))
	assert.Equal(t, time.Duration(0), d.SLAAt(4)) // terminal has no SLA
}

func TestPipelineTemplate_Lookup(t *testing.T) {
	d, ok := domain.PipelineTemplate("item-restoration")
	require.True(t, ok)
	assert.Equal(t, "Item Restoration", d.Name)

	_, ok = domain.PipelineTemplate("nope")
	assert.False(t, ok)
}
