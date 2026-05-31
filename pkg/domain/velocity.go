package domain

// VelocityBand is a SKU's rolling sales-velocity tier, derived from pick scans.
type VelocityBand string

const (
	VelocityHot  VelocityBand = "hot"
	VelocityWarm VelocityBand = "warm"
	VelocityCold VelocityBand = "cold"
	VelocityDead VelocityBand = "dead"
)

// Pick-scan thresholds over the rolling 7-day window.
const (
	velocityHotPicks7d  = 7 // ≥ this many picks in 7d → hot
	velocityWarmPicks7d = 1 // ≥ this many picks in 7d → warm
)

// VelocityFromPicks bands a SKU by its pick-scan counts. picks7d drives hot/warm;
// with no picks this week a SKU is cold if it still moved in the last 30d, else dead.
func VelocityFromPicks(picks7d, picks30d int) VelocityBand {
	switch {
	case picks7d >= velocityHotPicks7d:
		return VelocityHot
	case picks7d >= velocityWarmPicks7d:
		return VelocityWarm
	case picks30d > 0:
		return VelocityCold
	default:
		return VelocityDead
	}
}
