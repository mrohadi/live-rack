package domain

import "time"

// CardAge is a card's current stage plus how long it has dwelt there. Callers
// compute Age (now - entered_stage_at) so this stays pure and clock-free.
type CardAge struct {
	StagePosition int
	Age           time.Duration
}

// Bottleneck names the stage most over its SLA. AgeingCount is how many cards in
// the stage breach SLA; OldestAge is the worst dwell time among them.
type Bottleneck struct {
	Position    int           `json:"position"`
	Name        string        `json:"name"`
	AgeingCount int           `json:"ageing_count"`
	OldestAge   time.Duration `json:"oldest_age"`
}

// DetectBottleneck returns the worst stage where cards breach their SLA, or nil
// when nothing is ageing. The worst stage has the most breaching cards; ties
// break toward the stage holding the oldest card. Stages with SLA 0 (terminal /
// parking) never bottleneck.
func DetectBottleneck(stages []StageDef, cards []CardAge) *Bottleneck {
	type acc struct {
		count  int
		oldest time.Duration
	}
	byStage := make(map[int]*acc, len(stages))

	for _, c := range cards {
		if c.StagePosition < 0 || c.StagePosition >= len(stages) {
			continue
		}
		sla := stages[c.StagePosition].SLA
		if sla <= 0 || c.Age <= sla {
			continue
		}
		a := byStage[c.StagePosition]
		if a == nil {
			a = &acc{}
			byStage[c.StagePosition] = a
		}
		a.count++
		if c.Age > a.oldest {
			a.oldest = c.Age
		}
	}

	var best *Bottleneck
	for pos, a := range byStage {
		cand := &Bottleneck{
			Position:    pos,
			Name:        stages[pos].Name,
			AgeingCount: a.count,
			OldestAge:   a.oldest,
		}
		if best == nil ||
			cand.AgeingCount > best.AgeingCount ||
			(cand.AgeingCount == best.AgeingCount && cand.OldestAge > best.OldestAge) {
			best = cand
		}
	}
	return best
}
