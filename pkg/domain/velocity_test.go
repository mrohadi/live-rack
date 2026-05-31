package domain_test

import (
	"testing"

	"github.com/live-rack/pkg/domain"
)

func TestVelocityFromPicks(t *testing.T) {
	cases := []struct {
		name              string
		picks7d, picks30d int
		want              domain.VelocityBand
	}{
		{"hot at threshold", 7, 30, domain.VelocityHot},
		{"hot above", 20, 40, domain.VelocityHot},
		{"warm at one", 1, 5, domain.VelocityWarm},
		{"warm mid", 6, 10, domain.VelocityWarm},
		{"cold: none this week, moved in 30d", 0, 3, domain.VelocityCold},
		{"dead: no picks in 30d", 0, 0, domain.VelocityDead},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.VelocityFromPicks(tc.picks7d, tc.picks30d); got != tc.want {
				t.Fatalf("VelocityFromPicks(%d,%d) = %q, want %q",
					tc.picks7d, tc.picks30d, got, tc.want)
			}
		})
	}
}
