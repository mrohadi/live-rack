package domain

// CycleCountStatus is the lifecycle state of a count session.
type CycleCountStatus string

const (
	CycleCountOpen      CycleCountStatus = "open"
	CycleCountCompleted CycleCountStatus = "completed"
)

// CountVariance is counted minus system on-hand for one line. Positive means a
// surplus (more physically present than recorded); negative means shrinkage.
func CountVariance(systemQty, countedQty int) int {
	return countedQty - systemQty
}
