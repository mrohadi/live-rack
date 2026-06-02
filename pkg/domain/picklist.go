package domain

// PickListStatus is the lifecycle state of a pick list.
type PickListStatus string

const (
	PickListOpen      PickListStatus = "open"
	PickListPicking   PickListStatus = "picking"
	PickListCompleted PickListStatus = "completed"
	PickListCancelled PickListStatus = "cancelled"
)

// Valid reports whether s is a known pick-list status.
func (s PickListStatus) Valid() bool {
	switch s {
	case PickListOpen, PickListPicking, PickListCompleted, PickListCancelled:
		return true
	default:
		return false
	}
}

// PickLineStatus is the state of one line on a pick list.
type PickLineStatus string

const (
	PickLinePending PickLineStatus = "pending"
	PickLinePicked  PickLineStatus = "picked"
	PickLineShort   PickLineStatus = "short"
)

// PickLineOutcome classifies a reported pick against the requested quantity.
// picked >= requested resolves to picked; anything less is a short pick.
func PickLineOutcome(requested, picked int) PickLineStatus {
	if picked >= requested {
		return PickLinePicked
	}
	return PickLineShort
}

// PickShortfall returns requested minus picked, floored at zero.
func PickShortfall(requested, picked int) int {
	if picked >= requested {
		return 0
	}
	return requested - picked
}
