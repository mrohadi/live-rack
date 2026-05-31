package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/domain"
)

func TestPlanFromString(t *testing.T) {
	assert.Equal(t, domain.PlanStarter, domain.PlanFromString("starter"))
	assert.Equal(t, domain.PlanGrowth, domain.PlanFromString("growth"))
	assert.Equal(t, domain.PlanEnterprise, domain.PlanFromString("enterprise"))
	assert.Equal(t, domain.PlanFree, domain.PlanFromString("free"))
	assert.Equal(t, domain.PlanFree, domain.PlanFromString("bogus"))
}

func TestEntitlementsFor(t *testing.T) {
	free := domain.EntitlementsFor(domain.PlanFree)
	assert.Equal(t, 1, free.MaxStores)
	growth := domain.EntitlementsFor(domain.PlanGrowth)
	assert.Equal(t, 10, growth.MaxStores)
	ent := domain.EntitlementsFor(domain.PlanEnterprise)
	assert.Equal(t, -1, ent.ScansPerMin) // unlimited
	assert.Equal(t, free, domain.EntitlementsFor(domain.Plan("unknown")))
}

func TestWithinLimit(t *testing.T) {
	assert.True(t, domain.WithinLimit(-1, 9_999), "unlimited")
	assert.True(t, domain.WithinLimit(3, 2))
	assert.False(t, domain.WithinLimit(3, 3))
	assert.False(t, domain.WithinLimit(0, 0), "gated")
}
