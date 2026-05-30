// Package events defines NATS subjects and payloads for live-rack.
package events

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ScanRecorded is published after a scan is validated and persisted.
type ScanRecorded struct {
	OrgID     uuid.UUID `json:"org_id"`
	StoreID   uuid.UUID `json:"store_id"`
	ZoneID    uuid.UUID `json:"zone_id"`
	ScannerID string    `json:"scanner_id"`
	SKU       string    `json:"sku"`
	Action    string    `json:"action"`
	Valid     bool      `json:"valid"`
	Reason    string    `json:"reason,omitempty"`
	TS        time.Time `json:"ts"`
}

const subjectScanRecorded = "%s.scan.recorded"

// ScanSubject returns the per-org scan.recorded subject.
func ScanSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectScanRecorded, orgID)
}

// ExtractOrgID pulls the org segment from a "{org}.{domain}.{action}" subject.
func ExtractOrgID(subject string) string {
	return strings.SplitN(subject, ".", 2)[0]
}
