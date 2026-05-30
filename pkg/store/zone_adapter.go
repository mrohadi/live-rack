package store

import (
	"github.com/live-rack/pkg/domain"
)

// AsDomainZone converts the sqlc-generated Zone row into the pure domain entity.
// Constraints are passed through as raw bytes; callers use domain.UnmarshalConstraints
// when they need the typed view.
func (z Zone) AsDomainZone() domain.Zone {
	return domain.Zone{
		ID:          z.ID,
		OrgID:       z.OrgID,
		StoreID:     z.StoreID,
		Name:        z.Name,
		Type:        domain.ZoneType(z.Type),
		X:           z.X,
		Y:           z.Y,
		Width:       z.Width,
		Height:      z.Height,
		Color:       z.Color,
		Capacity:    int(z.Capacity),
		Constraints: z.Constraints,
		CreatedAt:   z.CreatedAt,
		UpdatedAt:   z.UpdatedAt,
	}
}
