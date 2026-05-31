// Package consumer turns pos.sale NATS messages into sales_events rows.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/store"
)

// SaleRecorder persists one sale within its org's RLS context.
type SaleRecorder interface {
	RecordSale(ctx context.Context, orgID uuid.UUID, arg store.CreateSaleEventParams) error
}

// Consumer handles decoded pos.sale messages.
type Consumer struct {
	rec SaleRecorder
}

// New builds a Consumer.
func New(rec SaleRecorder) *Consumer {
	return &Consumer{rec: rec}
}

// DecodeSale maps a pos.sale payload to insert params. Pure.
func DecodeSale(data []byte) (uuid.UUID, store.CreateSaleEventParams, error) {
	var s events.POSSale
	if err := json.Unmarshal(data, &s); err != nil {
		return uuid.Nil, store.CreateSaleEventParams{}, fmt.Errorf("ingest: decode pos.sale: %w", err)
	}
	if s.OrgID == uuid.Nil {
		return uuid.Nil, store.CreateSaleEventParams{}, fmt.Errorf("ingest: pos.sale missing org_id")
	}
	return s.OrgID, store.CreateSaleEventParams{
		Ts:          s.OccurredAt,
		OrgID:       s.OrgID,
		StoreID:     pgtype.UUID{Valid: false},
		Source:      s.Source,
		OrderID:     s.OrderID,
		Sku:         s.SKU,
		Qty:         s.Qty,
		AmountCents: s.AmountCents,
		Currency:    s.Currency,
		Channel:     s.Channel,
	}, nil
}

// Handle decodes a message and records the sale.
func (c *Consumer) Handle(ctx context.Context, data []byte) error {
	orgID, arg, err := DecodeSale(data)
	if err != nil {
		return err
	}
	if err := c.rec.RecordSale(ctx, orgID, arg); err != nil {
		return fmt.Errorf("ingest: record sale: %w", err)
	}
	return nil
}
