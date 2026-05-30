package store

import "github.com/live-rack/pkg/domain"

func (s ScanEvent) AsDomainScanEvent() domain.ScanEvent {
	return domain.ScanEvent{
		ID:        s.ID,
		Ts:        s.Ts,
		OrgID:     s.OrgID,
		StoreID:   s.StoreID,
		ZoneID:    s.ZoneID,
		ScannerID: s.ScannerID,
		SKU:       s.Sku,
		Action:    domain.ScanAction(s.Action),
		Valid:     s.Valid,
		Reason:    s.Reason,
	}
}
