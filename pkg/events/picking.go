package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PickProgress is published when a pick line is confirmed, so map/inventory
// clients can live-update on-hand counts and pick-run progress.
type PickProgress struct {
	OrgID     uuid.UUID `json:"org_id"`
	StoreID   uuid.UUID `json:"store_id"`
	ListID    uuid.UUID `json:"list_id"`
	LineID    uuid.UUID `json:"line_id"`
	SKU       string    `json:"sku"`
	QtyPicked int       `json:"qty_picked"`
	Status    string    `json:"status"`
	Done      int       `json:"done"`
	Total     int       `json:"total"`
	TS        time.Time `json:"ts"`
}

const subjectPickProgress = "lr.%s.pick.progress"

// PickProgressSubject returns the per-org pick.progress subject.
func PickProgressSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectPickProgress, orgID)
}
