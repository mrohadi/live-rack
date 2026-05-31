package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/domain"
)

func TestRequiresMFA(t *testing.T) {
	assert.True(t, domain.RequiresMFA(domain.PermEditUsers))
	assert.True(t, domain.RequiresMFA(domain.PermManageIntegrations))
	assert.False(t, domain.RequiresMFA(domain.PermViewDashboards))
}

func TestCanWithMFA(t *testing.T) {
	adminNoMFA := &domain.Principal{Role: domain.RoleAdmin, MFAVerified: false}
	adminMFA := &domain.Principal{Role: domain.RoleAdmin, MFAVerified: true}

	// Sensitive permission requires a second factor.
	assert.False(t, adminNoMFA.CanWithMFA(domain.PermEditUsers))
	assert.True(t, adminMFA.CanWithMFA(domain.PermEditUsers))

	// Non-sensitive permission unaffected by MFA.
	assert.True(t, adminNoMFA.CanWithMFA(domain.PermViewDashboards))

	// Role still gates: manager can't edit users even with MFA.
	assert.False(t, (&domain.Principal{Role: domain.RoleManager, MFAVerified: true}).CanWithMFA(domain.PermEditUsers))
}
