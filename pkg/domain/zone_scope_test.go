package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/domain"
)

func TestCanAccessZone(t *testing.T) {
	z1, z2, z3 := uuid.New(), uuid.New(), uuid.New()

	unscoped := &domain.Principal{Role: domain.RoleManager}
	assert.True(t, unscoped.CanAccessZone(z1), "empty ZoneIDs = org-wide")

	scoped := &domain.Principal{Role: domain.RoleStaff, ZoneIDs: []uuid.UUID{z1, z2}}
	assert.True(t, scoped.CanAccessZone(z1))
	assert.True(t, scoped.CanAccessZone(z2))
	assert.False(t, scoped.CanAccessZone(z3))
}
