package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PipelineBottleneck is published when a pipeline stage breaches its SLA, so the
// notifications/insight layer can alert a manager.
type PipelineBottleneck struct {
	OrgID       uuid.UUID `json:"org_id"`
	StoreID     uuid.UUID `json:"store_id"`
	PipelineID  uuid.UUID `json:"pipeline_id"`
	StagePos    int       `json:"stage_position"`
	StageName   string    `json:"stage_name"`
	AgeingCount int       `json:"ageing_count"`
	OldestAgeS  int64     `json:"oldest_age_seconds"`
	TS          time.Time `json:"ts"`
}

const subjectPipelineBottleneck = "lr.%s.pipeline.bottleneck"

// PipelineBottleneckSubject returns the per-org pipeline.bottleneck subject.
func PipelineBottleneckSubject(orgID uuid.UUID) string {
	return fmt.Sprintf(subjectPipelineBottleneck, orgID)
}
