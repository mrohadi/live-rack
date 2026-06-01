package domain

// Plan tiers + Org.Plan are declared in org.go. This file adds the per-plan
// entitlements (quota limits) and helpers — the single source of truth for
// billing-gated capacity.

// Entitlements are the quota limits a plan grants. A zero limit gates the
// resource; -1 means unlimited.
type Entitlements struct {
	MaxStores    int `json:"max_stores"`
	MaxSeats     int `json:"max_seats"`
	ScansPerMin  int `json:"scans_per_min"`
	Integrations int `json:"integrations"`
}

// planEntitlements is the source of truth for per-plan quotas.
var planEntitlements = map[Plan]Entitlements{
	PlanFree:       {MaxStores: 1, MaxSeats: 3, ScansPerMin: 600, Integrations: 1},
	PlanStarter:    {MaxStores: 3, MaxSeats: 10, ScansPerMin: 3000, Integrations: 3},
	PlanGrowth:     {MaxStores: 10, MaxSeats: 50, ScansPerMin: 6000, Integrations: 10},
	PlanEnterprise: {MaxStores: -1, MaxSeats: -1, ScansPerMin: -1, Integrations: -1},
}

// PlanFromString normalises a plan string, defaulting unknown values to free. Pure.
func PlanFromString(s string) Plan {
	switch Plan(s) {
	case PlanStarter:
		return PlanStarter
	case PlanGrowth:
		return PlanGrowth
	case PlanEnterprise:
		return PlanEnterprise
	default:
		return PlanFree
	}
}

// EntitlementsFor returns the entitlements for a plan. Pure.
func EntitlementsFor(p Plan) Entitlements {
	if e, ok := planEntitlements[p]; ok {
		return e
	}
	return planEntitlements[PlanFree]
}

// WithinLimit reports whether a current count is allowed by a limit, treating
// -1 as unlimited. Pure.
func WithinLimit(limit, current int) bool {
	if limit < 0 {
		return true
	}
	return current < limit
}
