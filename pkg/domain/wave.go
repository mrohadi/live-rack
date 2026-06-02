package domain

import "github.com/google/uuid"

// WaveStatus is the lifecycle state of a pick wave.
type WaveStatus string

const (
	WaveOpen      WaveStatus = "open"
	WavePicking   WaveStatus = "picking"
	WaveCompleted WaveStatus = "completed"
	WaveCancelled WaveStatus = "cancelled"
)

// Valid reports whether s is a known wave status.
func (s WaveStatus) Valid() bool {
	switch s {
	case WaveOpen, WavePicking, WaveCompleted, WaveCancelled:
		return true
	default:
		return false
	}
}

// LineDemand is one order line's outstanding requirement for a SKU+zone stop.
type LineDemand struct {
	LineID    uuid.UUID
	Requested int
}

// LineFill is the quantity allocated to one order line from a merged pick.
type LineFill struct {
	LineID uuid.UUID
	Picked int
	Status PickLineStatus
}

// AllocatePick distributes a merged picked quantity back across member order
// lines in FIFO order: each line is filled up to its requested qty before the
// next receives any. Lines beyond the supply are short (picked 0). Pure.
func AllocatePick(picked int, lines []LineDemand) []LineFill {
	out := make([]LineFill, 0, len(lines))
	remaining := picked
	for _, l := range lines {
		give := 0
		if remaining > 0 {
			give = l.Requested
			if remaining < give {
				give = remaining
			}
			remaining -= give
		}
		out = append(out, LineFill{
			LineID: l.LineID,
			Picked: give,
			Status: PickLineOutcome(l.Requested, give),
		})
	}
	return out
}
