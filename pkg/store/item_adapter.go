package store

import "github.com/live-rack/pkg/domain"

func (i Item) AsDomainItem() domain.Item {
	return domain.Item{
		ID:        i.ID,
		OrgID:     i.OrgID,
		SKU:       i.Sku,
		Name:      i.Name,
		Category:  i.Category,
		Status:    domain.ItemStatus(i.Status),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

func (l ItemLocation) AsDomainItemLocation() domain.ItemLocation {
	return domain.ItemLocation{
		ID:        l.ID,
		OrgID:     l.OrgID,
		StoreID:   l.StoreID,
		ZoneID:    l.ZoneID,
		SKU:       l.Sku,
		Qty:       int(l.Qty),
		UpdatedAt: l.UpdatedAt,
	}
}
