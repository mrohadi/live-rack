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
	ClerkUserID string    `json:"clerk_user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Principal carries verified identity after auth middleware.
type Principal struct {
	UserID     uuid.UUID
	OrgID      uuid.UUID
	ClerkOrgID string
	Role       RoleName
	StoreIDs   []uuid.UUID // empty = all stores
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
