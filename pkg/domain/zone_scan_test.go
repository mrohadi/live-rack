package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/live-rack/pkg/domain"
)

func TestZone_ValidateScan_Dwell(t *testing.T) {
	dwell := 60
	zone := domain.Zone{
		Constraints: mustConstraints(t, domain.ZoneConstraints{DwellSeconds: &dwell}),
	}
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		last    time.Time
		wantErr error
	}{
		{name: "no prior scan — accepts", last: time.Time{}},
		{name: "within dwell window — blocked", last: now.Add(-30 * time.Second), wantErr: domain.ErrDwellViolation},
		{name: "exactly at window — accepts", last: now.Add(-60 * time.Second)},
		{name: "past window — accepts", last: now.Add(-90 * time.Second)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := zone.ValidateScan(domain.ScanRequest{
				ScanQty: 1, LastScanAt: tc.last, Now: now,
			})
			assertErr(t, err, tc.wantErr)
		})
	}
}

func TestZone_ValidateScan_DualScan(t *testing.T) {
	zone := domain.Zone{
		Constraints: mustConstraints(t, domain.ZoneConstraints{RequireDualScan: true}),
	}
	now := time.Now()

	t.Run("unconfirmed — requires second scan", func(t *testing.T) {
		err := zone.ValidateScan(domain.ScanRequest{ScanQty: 1, Now: now})
		assertErr(t, err, domain.ErrDualScanRequired)
	})
	t.Run("confirmed — accepts", func(t *testing.T) {
		err := zone.ValidateScan(domain.ScanRequest{ScanQty: 1, Now: now, DualScanConfirmed: true})
		assertErr(t, err, nil)
	})
}

func assertErr(t *testing.T, got, want error) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
		return
	}
	if !errors.Is(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
