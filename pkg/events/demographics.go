package events

import (
	"fmt"

	"github.com/google/uuid"
)

// Demographics is published when a demographic snapshot is loaded for a store
// (or a specific zone within it). The ingest worker writes it into the
// ClickHouse demographics table for cross-signal analytics. Metric examples:
// "median_income", "foot_traffic", "age_25_34"; Segment scopes the metric
// (e.g. a catchment area or daypart).
type Demographics struct {
	OrgID   uuid.UUID `json:"org_id"`
	StoreID uuid.UUID `json:"store_id"`
	ZoneID  uuid.UUID `json:"zone_id"`
	Segment string    `json:"segment"`
	Metric  string    `json:"metric"`
	Value   float64   `json:"value"`
	Day     string    `json:"day"` // YYYY-MM-DD
}

const subjectDemographics = "lr.%s.demographics.snapshot"

// DemographicsSubject returns the per-org demographics.snapshot subject.
func DemographicsSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectDemographics, orgID)
}
