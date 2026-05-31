package domain

import (
	"encoding/json"
	"errors"
	"slices"
)

// Sentinel errors for ZoneConstraints validation
var (
	ErrDuplicateCategory = errors.New("zone constraints: duplicate category")
	ErrCategoryOverlap   = errors.New("zone constraints: category appear in both allowed and deined")
	ErrInvalidMaxPerSKU  = errors.New("zone constraints: max_units_per_sku must be >= 0")
	ErrEmptyCategory     = errors.New("zone constraints: category must be non-empty")
	ErrInvalidDwell      = errors.New("zone constraints: dwell_seconds must be >= 0")
)

// ZoneConstrains is the shaped type stored in zones.constraints JSON.
// A nil/zero value is valid (no constraints appleid).
type ZoneConstraints struct {
	AllowedCategories []string `json:"allowed_categories,omitempty"`
	DeniedCategories  []string `json:"denied_categories,omitempty"`
	MaxUnitsPerSKU    *int     `json:"max_units_per_sku,omitempty"`
	RequireDualScan   bool     `json:"require_dual_scan,omitempty"`
	DwellSeconds      *int     `json:"dwell_seconds,omitempty"`
}

// Validate returns nil if the constraints are internally consistent.
// STUB - to be implemented in step 2
func (c *ZoneConstraints) Validate() error {
	if err := validateCategoryList(c.AllowedCategories); err != nil {
		return err
	}
	if err := validateCategoryList(c.DeniedCategories); err != nil {
		return err
	}
	// Overlap check between allowed and denied.
	if len(c.AllowedCategories) > 0 && len(c.DeniedCategories) > 0 {
		allowed := make(map[string]struct{}, len(c.AllowedCategories))
		for _, cat := range c.AllowedCategories {
			allowed[cat] = struct{}{}
		}
		for _, cat := range c.DeniedCategories {
			if _, ok := allowed[cat]; ok {
				return ErrCategoryOverlap
			}
		}
	}
	if c.MaxUnitsPerSKU != nil && *c.MaxUnitsPerSKU < 0 {
		return ErrInvalidMaxPerSKU
	}
	if c.DwellSeconds != nil && *c.DwellSeconds < 0 {
		return ErrInvalidDwell
	}

	return nil
}

// validateCategoryList rejects empty strings and duplicates within a single list
func validateCategoryList(cats []string) error {
	seen := make(map[string]struct{}, len(cats))
	for _, cat := range cats {
		if cat == "" {
			return ErrEmptyCategory
		}
		if _, dup := seen[cat]; dup {
			return ErrDuplicateCategory
		}

		seen[cat] = struct{}{}
	}

	return nil
}

// Sentinel errors for Zone.CanAcceptItem decisions.
var (
	ErrCategoryNotAllowed = errors.New("zone: item category not in allowed list")
	ErrCategoryDenied     = errors.New("zone: item category is denied")
	ErrCapacityExceeded   = errors.New("zone: capacity exceeded")
	ErrMaxPerSKUExceeded  = errors.New("zone: max units per SKU exceeded")
	ErrInvalidScanQty     = errors.New("zone: scan quantity must be > 0")
)

// MarshalConstraints encodes a typed ZoneConstraints int the []byte form
// stored on Zone.Constraints. Exposed for test setup and API layer use.
func MarshalConstraints(c ZoneConstraints) ([]byte, error) {
	return json.Marshal(c)
}

// UnmarshalConstraints decodes the bytes stored on Zone.Constraints.
// Empty/nil bytes decode to a zero-value (no constraints)
func UnmarshalConstraints(b []byte) (ZoneConstraints, error) {
	var c ZoneConstraints
	if len(b) == 0 {
		return c, nil
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, nil
	}
	return c, nil
}

// CanAcceptItem decides whether the zone may accept scanQty units of itemCategory,
// given currentQty already present (of that SKU, for the per-SKU check; total for capacity).
// STUB - to be implemented in step 3c.
func (z Zone) CanAcceptItem(itemCategory string, currentQty int, scanQty int) error {
	if scanQty <= 0 {
		return ErrInvalidScanQty
	}

	c, err := UnmarshalConstraints(z.Constraints)
	if err != nil {
		// Treat malformed stored constraints as "no constraints" rather than
		// blocking scans; bad data is surfaced via the API validate path on write.
		c = ZoneConstraints{}
	}

	// Denied list takes precedence over allowed.
	if itemCategory != "" {
		if slices.Contains(c.DeniedCategories, itemCategory) {
			return ErrCategoryDenied
		}
		if len(c.AllowedCategories) > 0 {
			ok := slices.Contains(c.AllowedCategories, itemCategory)
			if !ok {
				return ErrCategoryNotAllowed
			}
		}
	}

	// Capacity = 0 means unlimited (matches existing DB default).
	if z.Capacity > 0 && currentQty+scanQty > z.Capacity {
		return ErrCapacityExceeded
	}

	if c.MaxUnitsPerSKU != nil && currentQty+scanQty > *c.MaxUnitsPerSKU {
		return ErrMaxPerSKUExceeded
	}

	return nil
}
