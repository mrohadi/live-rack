package domain

import (
	"time"

	"github.com/google/uuid"
)

type Plan string

const (
	PlanFree       Plan = "free"
	PlanStarter    Plan = "starter"
	PlanGrowth     Plan = "growth"
	PlanEnterprise Plan = "enterprise"
)

type Org struct {
	ID        uuid.UUID `json:"id"`
	IDPOrgID  string    `json:"idp_org_id"`
	Name      string    `json:"name"`
	Plan      Plan      `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address,omitempty"`
	Lat       float64   `json:"lat,omitempty"`
	Lon       float64   `json:"lon,omitempty"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
