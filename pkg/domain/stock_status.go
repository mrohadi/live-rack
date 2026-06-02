package domain

// StockStatus bands an item location by on-hand quantity relative to its
// reorder point. Distinct from ItemStatus, which is catalog lifecycle.
type StockStatus string

const (
	StockStatusInStock StockStatus = "in_stock"
	StockStatusLow     StockStatus = "low"
	StockStatusOut     StockStatus = "out"
)

// StockStatusFromQty bands on-hand qty against a reorder point.
//
//	qty <= 0                          → out
//	0 < qty <= reorderPoint (rp > 0)  → low
//	otherwise                         → in_stock
//
// A reorder point of 0 disables the "low" band: any positive qty is in_stock.
func StockStatusFromQty(qty, reorderPoint int) StockStatus {
	switch {
	case qty <= 0:
		return StockStatusOut
	case reorderPoint > 0 && qty <= reorderPoint:
		return StockStatusLow
	default:
		return StockStatusInStock
	}
}

// NeedsReorder reports whether on-hand qty has fallen to/below the reorder
// point. False when reorderPoint is 0 (trigger disabled).
func NeedsReorder(qty, reorderPoint int) bool {
	return reorderPoint > 0 && qty <= reorderPoint
}
