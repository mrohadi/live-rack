package domain

// ShipmentStatus is the lifecycle state of an outbound shipment.
type ShipmentStatus string

const (
	ShipmentPacking    ShipmentStatus = "packing"
	ShipmentPacked     ShipmentStatus = "packed"
	ShipmentDispatched ShipmentStatus = "dispatched"
	ShipmentCancelled  ShipmentStatus = "cancelled"
)

// Valid reports whether s is a known shipment status.
func (s ShipmentStatus) Valid() bool {
	switch s {
	case ShipmentPacking, ShipmentPacked, ShipmentDispatched, ShipmentCancelled:
		return true
	default:
		return false
	}
}

// CanDispatch reports whether a shipment in the given status may be dispatched.
// Only a packed shipment ships. Pure.
func CanDispatch(s ShipmentStatus) bool {
	return s == ShipmentPacked
}

// CanCancelShipment reports whether a shipment may still be cancelled (anything
// before it leaves the building). Pure.
func CanCancelShipment(s ShipmentStatus) bool {
	return s == ShipmentPacking || s == ShipmentPacked
}
