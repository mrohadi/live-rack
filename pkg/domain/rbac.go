package domain

import "sort"

// Permission is a single capability checked across the product. The matrix
// below is the single source of truth, mirroring the design's role grid; the
// scattered Can* helpers delegate here.
type Permission string

const (
	PermViewDashboards     Permission = "view_dashboards"
	PermEditZones          Permission = "edit_zones"
	PermApproveMisscans    Permission = "approve_misscans"
	PermManagePipelines    Permission = "manage_pipelines"
	PermRunScanner         Permission = "run_scanner"
	PermMoveInventory      Permission = "move_inventory"
	PermManageTasksAny     Permission = "manage_tasks_any"
	PermManageTasksOwn     Permission = "manage_tasks_own"
	PermEditUsers          Permission = "edit_users"
	PermManageIntegrations Permission = "manage_integrations"
	PermExportReports      Permission = "export_reports"
)

// rolePermissions is the role × permission matrix. A missing entry denies.
var rolePermissions = map[RoleName]map[Permission]bool{
	RoleAdmin: {
		PermViewDashboards: true, PermEditZones: true, PermApproveMisscans: true,
		PermManagePipelines: true, PermRunScanner: true, PermMoveInventory: true,
		PermManageTasksAny: true, PermManageTasksOwn: true, PermEditUsers: true,
		PermManageIntegrations: true, PermExportReports: true,
	},
	RoleManager: {
		PermViewDashboards: true, PermEditZones: true, PermApproveMisscans: true,
		PermManagePipelines: true, PermRunScanner: true, PermMoveInventory: true,
		PermManageTasksAny: true, PermManageTasksOwn: true, PermExportReports: true,
	},
	RoleStaff: {
		PermViewDashboards: true, PermRunScanner: true, PermMoveInventory: true,
		PermManageTasksOwn: true,
	},
	RoleReadonly: {
		PermViewDashboards: true, PermExportReports: true,
	},
	RoleService: {
		PermRunScanner: true, PermMoveInventory: true,
	},
}

// mfaRequired are high-impact permissions that demand a second factor.
var mfaRequired = map[Permission]bool{
	PermEditUsers:          true,
	PermManageIntegrations: true,
}

// RequiresMFA reports whether a permission demands a verified second factor. Pure.
func RequiresMFA(perm Permission) bool { return mfaRequired[perm] }

// Can reports whether a role holds a permission. Pure.
func Can(role RoleName, perm Permission) bool {
	return rolePermissions[role][perm]
}

// CanWithMFA reports whether the principal may exercise a permission, enforcing
// a second factor for MFA-required permissions. Pure.
func (p *Principal) CanWithMFA(perm Permission) bool {
	if !Can(p.Role, perm) {
		return false
	}
	return !RequiresMFA(perm) || p.MFAVerified
}

// Permissions returns the sorted permissions granted to a role. Pure.
func Permissions(role RoleName) []Permission {
	set := rolePermissions[role]
	out := make([]Permission, 0, len(set))
	for p, ok := range set {
		if ok {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
