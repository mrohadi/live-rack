package domain_test

import (
	"testing"

	"github.com/live-rack/pkg/domain"
)

func TestStockStatusFromQty(t *testing.T) {
	cases := []struct {
		name         string
		qty          int
		reorderPoint int
		want         domain.StockStatus
	}{
		{"out at zero", 0, 5, domain.StockStatusOut},
		{"out when negative", -2, 5, domain.StockStatusOut},
		{"low at reorder point", 5, 5, domain.StockStatusLow},
		{"low below reorder point", 3, 5, domain.StockStatusLow},
		{"in stock above reorder point", 6, 5, domain.StockStatusInStock},
		{"in stock when reorder disabled", 1, 0, domain.StockStatusInStock},
		{"out when reorder disabled but empty", 0, 0, domain.StockStatusOut},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.StockStatusFromQty(tc.qty, tc.reorderPoint); got != tc.want {
				t.Fatalf("StockStatusFromQty(%d,%d) = %q, want %q",
					tc.qty, tc.reorderPoint, got, tc.want)
			}
		})
	}
}

func TestNeedsReorder(t *testing.T) {
	cases := []struct {
		name         string
		qty          int
		reorderPoint int
		want         bool
	}{
		{"at point triggers", 5, 5, true},
		{"below point triggers", 2, 5, true},
		{"above point no trigger", 6, 5, false},
		{"disabled never triggers", 0, 0, false},
		{"empty with point triggers", 0, 3, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.NeedsReorder(tc.qty, tc.reorderPoint); got != tc.want {
				t.Fatalf("NeedsReorder(%d,%d) = %v, want %v",
					tc.qty, tc.reorderPoint, got, tc.want)
			}
		})
	}
}
