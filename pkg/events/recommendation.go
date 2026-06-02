package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Recommendation is published by the insight engine when a signal warrants an
// operator action. The UI surfaces it on the Analytics screen with an Apply
// button that turns SuggestedTask into a real task.
type Recommendation struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"org_id"`
	StoreID       uuid.UUID `json:"store_id"`
	Kind          string    `json:"kind"`     // "weather" | "transit"
	Severity      string    `json:"severity"` // "info" | "action"
	Title         string    `json:"title"`
	Rationale     string    `json:"rationale"`
	SuggestedTask string    `json:"suggested_task"`
	CreatedAt     time.Time `json:"created_at"`
}

const subjectRecommendation = "lr.%s.recommendation.created"

// RecommendationSubject returns the per-org recommendation.created subject.
func RecommendationSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectRecommendation, orgID)
}
