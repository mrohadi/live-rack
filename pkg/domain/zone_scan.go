package domain

import (
	"errors"
	"time"
)

var (
	ErrDwellViolation   = errors.New("zone: item rescanned inside dwell window")
	ErrDualScanRequired = errors.New("zone: dual scan confirmation required")
)

// ScanRequest carries the runtime state a single scan validates against.
type ScanRequest struct {
	Category          string
	CurrentQty        int
	ScanQty           int
	LastScanAt        time.Time // zero = no prior scan of this SKU in this zone
	Now               time.Time
	DualScanConfirmed bool
}

// ValidateScan runs the full mis-scan rule chain: category, capacity,
// max-per-SKU, dwell, then dual-scan.
func (z Zone) ValidateScan(req ScanRequest) error {
	if err := z.CanAcceptItem(req.Category, req.CurrentQty, req.ScanQty); err != nil {
		return err
	}

	c, err := UnmarshalConstraints(z.Constraints)
	if err != nil {
		c = ZoneConstraints{}
	}

	if c.DwellSeconds != nil && !req.LastScanAt.IsZero() {
		window := time.Duration(*c.DwellSeconds) * time.Second
		if req.Now.Sub(req.LastScanAt) < window {
			return ErrDwellViolation
		}
	}

	if c.RequireDualScan && !req.DualScanConfirmed {
		return ErrDualScanRequired
	}

	return nil
}
