package domain

import "time"

// Built-in pipeline template keys.
const (
	TemplateItemRestoration = "item-restoration"
)

// RestorationTemplate is the first-class pipeline wedge: non-linear refurb flow
// Intake → Triage → Repair → QA → Restocked. SLAs match the design bundle.
func RestorationTemplate() PipelineDef {
	return PipelineDef{
		Key:  TemplateItemRestoration,
		Name: "Item Restoration",
		Stages: []StageDef{
			{Name: "Intake", SLA: 24 * time.Hour},
			{Name: "Triage", SLA: 48 * time.Hour},
			{Name: "Repair", SLA: 5 * 24 * time.Hour},
			{Name: "QA", SLA: 24 * time.Hour},
			{Name: "Restocked", SLA: 0}, // terminal
		},
	}
}

// pipelineTemplates is the registry of built-in templates keyed by Key.
var pipelineTemplates = map[string]func() PipelineDef{
	TemplateItemRestoration: RestorationTemplate,
}

// PipelineTemplate returns the named built-in template, or ok=false if unknown.
func PipelineTemplate(key string) (PipelineDef, bool) {
	build, ok := pipelineTemplates[key]
	if !ok {
		return PipelineDef{}, false
	}
	return build(), true
}
