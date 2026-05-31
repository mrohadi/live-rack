package domain

import (
	"time"

	"github.com/google/uuid"
)

// ItemStatus is the catalog lifecycle state of an item.
type ItemStatus string

const (
	ItemStatusActive       ItemStatus = "active"
	ItemStatusDiscontinued ItemStatus = "discontinued"
	ItemStatusRecalled     ItemStatus = "recalled"
)

// Item is a master-catalog product, unique per org by SKU.
type Item struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	SKU       string     `json:"sku"`
	Name      string     `json:"name"`
	Category  string     `json:"category"`
	Status    ItemStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ItemLocation is the current on-hand quantity of a SKU in one zone.
type ItemLocation struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	StoreID   uuid.UUID `json:"store_id"`
	ZoneID    uuid.UUID `json:"zone_id"`
	SKU       string    `json:"sku"`
	Qty       int       `json:"qty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QtyDelta returns the signed quantity change a scan action applies to a
// location: place/count/move add stock, pick removes it.
func QtyDelta(action ScanAction, qty int) int {
	switch action {
	case ScanActionPick:
		return -qty
	default:
		return qty
	}
}
