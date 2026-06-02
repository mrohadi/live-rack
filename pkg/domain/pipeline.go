package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPipelineKeyRequired  = errors.New("pipeline: key must be non-empty")
	ErrPipelineNameRequired = errors.New("pipeline: name must be non-empty")
	ErrPipelineNoStages     = errors.New("pipeline: at least one stage required")
	ErrStageNameRequired    = errors.New("pipeline: stage name must be non-empty")
	ErrStageNegativeSLA     = errors.New("pipeline: stage SLA must be >= 0")
	ErrStageDuplicate       = errors.New("pipeline: duplicate stage name")
)

// CardPriority orders cards within a stage. Mirrors the design bundle vocabulary
// (low/medium/high), distinct from TaskPriority which uses "med".
type CardPriority string

const (
	CardPriorityLow    CardPriority = "low"
	CardPriorityMedium CardPriority = "medium"
	CardPriorityHigh   CardPriority = "high"
)

func (p CardPriority) Valid() bool {
	switch p {
	case CardPriorityLow, CardPriorityMedium, CardPriorityHigh:
		return true
	default:
		return false
	}
}

// StageDef is one ordered step in a pipeline template. SLA is the target dwell
// time for a card in this stage; zero means no SLA (terminal/parking stages).
type StageDef struct {
	Name string        `json:"name"`
	SLA  time.Duration `json:"sla"`
}

// PipelineDef is the per-org pipeline DSL: an ordered list of named stages with
// SLAs. It is a pure value — instantiating it produces a Pipeline + PipelineStages.
type PipelineDef struct {
	Key    string     `json:"key"`
	Name   string     `json:"name"`
	Stages []StageDef `json:"stages"`
}

// Validate reports the first structural problem with a definition, or nil.
func (d PipelineDef) Validate() error {
	if strings.TrimSpace(d.Key) == "" {
		return ErrPipelineKeyRequired
	}
	if strings.TrimSpace(d.Name) == "" {
		return ErrPipelineNameRequired
	}
	if len(d.Stages) == 0 {
		return ErrPipelineNoStages
	}
	seen := make(map[string]struct{}, len(d.Stages))
	for _, s := range d.Stages {
		if strings.TrimSpace(s.Name) == "" {
			return ErrStageNameRequired
		}
		if s.SLA < 0 {
			return ErrStageNegativeSLA
		}
		key := strings.ToLower(strings.TrimSpace(s.Name))
		if _, dup := seen[key]; dup {
			return ErrStageDuplicate
		}
		seen[key] = struct{}{}
	}
	return nil
}

// StageCount returns the number of stages.
func (d PipelineDef) StageCount() int { return len(d.Stages) }

// StageIndex returns the zero-based position of the named stage, or -1.
// Matching is case-insensitive and trims surrounding whitespace.
func (d PipelineDef) StageIndex(name string) int {
	want := strings.ToLower(strings.TrimSpace(name))
	for i, s := range d.Stages {
		if strings.ToLower(strings.TrimSpace(s.Name)) == want {
			return i
		}
	}
	return -1
}

// IsTerminalStage reports whether position is the final stage (cards leave here).
func (d PipelineDef) IsTerminalStage(position int) bool {
	return position == len(d.Stages)-1
}

// SLAAt returns the SLA for the stage at position, or zero if out of range.
func (d PipelineDef) SLAAt(position int) time.Duration {
	if position < 0 || position >= len(d.Stages) {
		return 0
	}
	return d.Stages[position].SLA
}

// Pipeline is an instantiated workflow scoped to a store.
type Pipeline struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	StoreID   uuid.UUID `json:"store_id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CanMutatePipeline reports whether the principal may create pipelines, edit
// stages, or move cards. Matches the design permission matrix: Admin + Manager.
func CanMutatePipeline(p *Principal) bool {
	return Can(p.Role, PermManagePipelines)
}
