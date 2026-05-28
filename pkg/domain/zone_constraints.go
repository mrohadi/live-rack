package domain

import "errors"

// Sentinel errors for ZoneConstraints validation
var (
	ErrDuplicateCategory = errors.New("zone constraints: duplicate category")
	ErrCategoryOverlap   = errors.New("zone constraints: category appear in both allowed and deined")
	ErrInvalidMaxPerSKU  = errors.New("zone constraints: max_units_per_sku must be >= 0")
	ErrEmptyCategory     = errors.New("zone constraints: category must be non-empty")
)

// ZoneConstrains is the shaped type stored in zones.constraints JSON.
// A nil/zero value is valid (no constraints appleid).
type ZoneConstraints struct {
	AllowedCategories []string `json:"allowed_categories,omitempty"`
	DeniedCategories  []string `json:"denied_categories,omitempty"`
	MaxUnitsPerSKU    *int     `json:"max_units_per_sku,omitempty"`
	RequireDualScan   bool     `json:"require_dual_scan,omitempty"`
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
