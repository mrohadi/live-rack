package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/domain"
)

func TestCan_Matrix(t *testing.T) {
	cases := []struct {
		role domain.RoleName
		perm domain.Permission
		want bool
	}{
		{domain.RoleAdmin, domain.PermEditUsers, true},
		{domain.RoleManager, domain.PermEditUsers, false},
		{domain.RoleManager, domain.PermEditZones, true},
		{domain.RoleStaff, domain.PermRunScanner, true},
		{domain.RoleStaff, domain.PermManageTasksAny, false},
		{domain.RoleStaff, domain.PermManageTasksOwn, true},
		{domain.RoleReadonly, domain.PermViewDashboards, true},
		{domain.RoleReadonly, domain.PermExportReports, true},
		{domain.RoleReadonly, domain.PermRunScanner, false},
		{domain.RoleService, domain.PermRunScanner, true},
		{domain.RoleService, domain.PermViewDashboards, false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, domain.Can(c.role, c.perm), "%s / %s", c.role, c.perm)
	}
}

func TestPermissions_SortedAndScoped(t *testing.T) {
	admin := domain.Permissions(domain.RoleAdmin)
	assert.Len(t, admin, 11)
	// sorted ascending
	for i := 1; i < len(admin); i++ {
		assert.Less(t, string(admin[i-1]), string(admin[i]))
	}

	ro := domain.Permissions(domain.RoleReadonly)
	assert.Equal(t, []domain.Permission{domain.PermExportReports, domain.PermViewDashboards}, ro)

	assert.Empty(t, domain.Permissions(domain.RoleName("bogus")))
}

func TestCanHelpers_DelegateToMatrix(t *testing.T) {
	assert.True(t, domain.CanManageIntegrations(&domain.Principal{Role: domain.RoleAdmin}))
	assert.False(t, domain.CanManageIntegrations(&domain.Principal{Role: domain.RoleManager}))
	assert.True(t, domain.CanMutatePipeline(&domain.Principal{Role: domain.RoleManager}))
	assert.False(t, domain.CanMutatePipeline(&domain.Principal{Role: domain.RoleStaff}))
	assert.True(t, domain.CanMutateTask(&domain.Principal{Role: domain.RoleStaff}))
	assert.False(t, domain.CanMutateTask(&domain.Principal{Role: domain.RoleReadonly}))
}
