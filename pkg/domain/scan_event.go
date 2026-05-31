package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScanAction string

const (
	ScanActionPlace ScanAction = "place"
	ScanActionPick  ScanAction = "pick"
	ScanActionMove  ScanAction = "move"
	ScanActionCount ScanAction = "count"
)

type ScanEvent struct {
	ID        uuid.UUID  `json:"id"`
	Ts        time.Time  `json:"ts"`
	OrgID     uuid.UUID  `json:"org_id"`
	StoreID   uuid.UUID  `json:"store_id"`
	ZoneID    uuid.UUID  `json:"zone_id"`
	ScannerID string     `json:"scanner_id"`
	SKU       string     `json:"sku"`
	Action    ScanAction `json:"action"`
	Valid     bool       `json:"valid"`
	Reason    string     `json:"reason"`
}
