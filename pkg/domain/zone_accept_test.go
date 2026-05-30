package domain_test

import (
	"errors"
	"testing"

	"github.com/live-rack/pkg/domain"
)

func TestZone_CanAcceptItem(t *testing.T) {
	max5 := 5

	cases := []struct {
		name       string
		zone       domain.Zone
		category   string
		currentQty int
		scanQty    int
		wantErr    error
	}{
		{
			name:    "no constraints, no capacity — accepts",
			zone:    domain.Zone{},
			scanQty: 1,
		},
		{
			name: "category in allowed list — accepts",
			zone: domain.Zone{
				Constraints: mustConstraints(t, domain.ZoneConstraints{
					AllowedCategories: []string{"frozen", "beverage"},
				}),
			},
			category: "frozen",
			scanQty:  1,
		},
		{
			name: "category not in allowed list — denied",
			zone: domain.Zone{
				Constraints: mustConstraints(t, domain.ZoneConstraints{
					AllowedCategories: []string{"frozen"},
				}),
			},
			category: "hazmat",
			scanQty:  1,
			wantErr:  domain.ErrCategoryNotAllowed,
		},
		{
			name: "category in denied list — denied",
			zone: domain.Zone{
				Constraints: mustConstraints(t, domain.ZoneConstraints{
					DeniedCategories: []string{"hazmat"},
				}),
			},
			category: "hazmat",
			scanQty:  1,
			wantErr:  domain.ErrCategoryDenied,
		},
		{
			name:       "capacity exceeded by scan",
			zone:       domain.Zone{Capacity: 10},
			currentQty: 8,
			scanQty:    5,
			wantErr:    domain.ErrCapacityExceeded,
		},
		{
			name:       "capacity exactly reached — accepts",
			zone:       domain.Zone{Capacity: 10},
			currentQty: 7,
			scanQty:    3,
		},
		{
			name:       "zero capacity means unlimited — accepts",
			zone:       domain.Zone{Capacity: 0},
			currentQty: 9999,
			scanQty:    1,
		},
		{
			name: "max per sku exceeded",
			zone: domain.Zone{
				Constraints: mustConstraints(t, domain.ZoneConstraints{
					MaxUnitsPerSKU: &max5,
				}),
			},
			currentQty: 4,
			scanQty:    2,
			wantErr:    domain.ErrMaxPerSKUExceeded,
		},
		{
			name: "max per sku exactly reached — accepts",
			zone: domain.Zone{
				Constraints: mustConstraints(t, domain.ZoneConstraints{
					MaxUnitsPerSKU: &max5,
				}),
			},
			currentQty: 3,
			scanQty:    2,
		},
		{
			name:    "non-positive scan quantity rejected",
			zone:    domain.Zone{},
			scanQty: 0,
			wantErr: domain.ErrInvalidScanQty,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.zone.CanAcceptItem(tc.category, tc.currentQty, tc.scanQty)
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

// mustConstraints marshall a typed ZoneConstraints into the []byte from zone holds.
func mustConstraints(t *testing.T, c domain.ZoneConstraints) []byte {
	t.Helper()
	b, err := domain.MarshalConstraints(c)
	if err != nil {
		t.Fatalf("marshal constraints: %v", err)
	}
	return b
}
