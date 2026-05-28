package domain_test

import (
	"errors"
	"testing"

	"github.com/live-rack/pkg/domain"
)

func TestZoneConstraints_Validate(t *testing.T) {
	maxNeg := -1
	maxOK := 10

	cases := []struct {
		name    string
		in      domain.ZoneConstraints
		wantErr error
	}{
		{
			name: "empty is valid",
			in:   domain.ZoneConstraints{},
		},
		{
			name: "allowed only is valid",
			in:   domain.ZoneConstraints{AllowedCategories: []string{"frozen", "beverage"}},
		},
		{
			name: "denied only is valid",
			in:   domain.ZoneConstraints{DeniedCategories: []string{"hazmat"}},
		},
		{
			name:    "duplicate allowed category",
			in:      domain.ZoneConstraints{AllowedCategories: []string{"frozen", "frozen"}},
			wantErr: domain.ErrDuplicateCategory,
		},
		{
			name:    "duplicate denied category",
			in:      domain.ZoneConstraints{DeniedCategories: []string{"hazmat", "hazmat"}},
			wantErr: domain.ErrDuplicateCategory,
		},
		{
			name: "overlap between allowed and denied",
			in: domain.ZoneConstraints{
				AllowedCategories: []string{"frozen"},
				DeniedCategories:  []string{"frozen"},
			},
			wantErr: domain.ErrCategoryOverlap,
		},
		{
			name:    "negative max per sku",
			in:      domain.ZoneConstraints{MaxUnitsPerSKU: &maxNeg},
			wantErr: domain.ErrInvalidMaxPerSKU,
		},
		{
			name: "valid max per sku",
			in:   domain.ZoneConstraints{MaxUnitsPerSKU: &maxOK},
		},
		{
			name:    "empty string category rejected",
			in:      domain.ZoneConstraints{AllowedCategories: []string{""}},
			wantErr: domain.ErrEmptyCategory,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("want nil, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
		})
	}
}
