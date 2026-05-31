package domain

import "testing"

func TestQtyDelta(t *testing.T) {
	cases := []struct {
		action ScanAction
		qty    int
		want   int
	}{
		{ScanActionPlace, 3, 3},
		{ScanActionCount, 5, 5},
		{ScanActionMove, 2, 2},
		{ScanActionPick, 4, -4},
	}
	for _, tc := range cases {
		if got := QtyDelta(tc.action, tc.qty); got != tc.want {
			t.Errorf("QtyDelta(%q, %d) = %d, want %d", tc.action, tc.qty, got, tc.want)
		}
	}
}
