package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// POSSale is published when a verified POS webhook yields a sale line. The ingest
// worker consumes it into the sales_events hypertable.
type POSSale struct {
	OrgID       uuid.UUID `json:"org_id"`
	Source      string    `json:"source"`
	OrderID     string    `json:"order_id"`
	SKU         string    `json:"sku"`
	Qty         int32     `json:"qty"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	Channel     string    `json:"channel"`
	OccurredAt  time.Time `json:"occurred_at"`
}

const subjectPOSSale = "lr.%s.pos.sale"

// POSSaleSubject returns the per-org pos.sale subject.
func POSSaleSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectPOSSale, orgID)
}
