package domain_test

import (
	"testing"

	"github.com/live-rack/pkg/domain"
)

func TestCountVariance(t *testing.T) {
	cases := []struct {
		name            string
		system, counted int
		want            int
	}{
		{"match", 10, 10, 0},
		{"shrinkage", 10, 7, -3},
		{"surplus", 5, 8, 3},
		{"counted zero", 4, 0, -4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.CountVariance(tc.system, tc.counted); got != tc.want {
				t.Fatalf("CountVariance(%d,%d) = %d, want %d", tc.system, tc.counted, got, tc.want)
			}
		})
	}
}
