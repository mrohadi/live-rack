package domain

import (
	"time"

	"github.com/google/uuid"
)

type ZoneType string

const (
	ZoneTypeGeneral  ZoneType = "general"
	ZoneTypeFrozen   ZoneType = "frozen"
	ZoneTypeReturns  ZoneType = "returns"
	ZoneTypeStaging  ZoneType = "staging"
	ZoneTypeDisplay  ZoneType = "display"
	ZoneTypeCheckout ZoneType = "checkout"
)

type Zone struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	StoreID     uuid.UUID `json:"store_id"`
	Name        string    `json:"name"`
	Type        ZoneType  `json:"type"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Width       float64   `json:"width"`
	Height      float64   `json:"height"`
	Color       string    `json:"color"`
	Capacity    int       `json:"capacity"`
	Constraints []byte    `json:"constraints"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
