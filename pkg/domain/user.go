package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoleName string

const (
	RoleAdmin    RoleName = "admin"
	RoleManager  RoleName = "manager"
	RoleStaff    RoleName = "staff"
	RoleReadonly RoleName = "readonly"
	RoleService  RoleName = "service"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	IDPUserID   string    `json:"idp_user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Principal carries verified identity after auth middleware.
type Principal struct {
	UserID      uuid.UUID
	IDPUserID   string // Zitadel subject — targets the user in IdP-side calls
	OrgID       uuid.UUID
	IDPOrgID    string
	Role        RoleName
	StoreIDs    []uuid.UUID // empty = all stores
	ZoneIDs     []uuid.UUID // empty = all zones (org-wide)
	MFAVerified bool        // a second factor was used this session
}

func (p *Principal) HasRole(roles ...RoleName) bool {
	for _, r := range roles {
		if p.Role == r {
			return true
		}
	}
	return false
}

func (p *Principal) CanAccessStore(storeID uuid.UUID) bool {
	if len(p.StoreIDs) == 0 {
		return true
	}
	for _, sid := range p.StoreIDs {
		if sid == storeID {
			return true
		}
	}
	return false
}

// CanAccessZone reports whether the principal may access a zone. An empty
// ZoneIDs set means org-wide access (admins, managers); otherwise access is
// restricted to the assigned zones. Mirrors the user_zones RLS predicate.
func (p *Principal) CanAccessZone(zoneID uuid.UUID) bool {
	if len(p.ZoneIDs) == 0 {
		return true
	}
	for _, zid := range p.ZoneIDs {
		if zid == zoneID {
			return true
		}
	}
	return false
}
