// Package chsink decodes scan.recorded and pos.sale NATS messages into
// ClickHouse raw-table rows and writes them via an Inserter. Decoding is pure
// so it can be unit-tested without a warehouse.
package chsink

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/live-rack/pkg/events"
)

const (
	scanTable = "scan_events_raw"
	saleTable = "sales_events_raw"
	// chTimeLayout matches ClickHouse DateTime64(3) text parsing.
	chTimeLayout = "2006-01-02 15:04:05.000"
)

// Inserter writes rows into a ClickHouse table. *chstore.Client satisfies it.
type Inserter interface {
	Insert(ctx context.Context, table string, rows []map[string]any) error
}

func boolToUInt8(b bool) int {
	if b {
		return 1
	}
	return 0
}

// DecodeScan maps a scan.recorded payload to a scan_events_raw row. Pure.
func DecodeScan(data []byte) (map[string]any, error) {
	var s events.ScanRecorded
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("chsink: decode scan.recorded: %w", err)
	}
	if s.OrgID == uuid.Nil {
		return nil, fmt.Errorf("chsink: scan.recorded missing org_id")
	}
	return map[string]any{
		"org_id":     s.OrgID.String(),
		"ts":         s.TS.UTC().Format(chTimeLayout),
		"store_id":   s.StoreID.String(),
		"zone_id":    s.ZoneID.String(),
		"scanner_id": s.ScannerID,
		"sku":        s.SKU,
		"action":     s.Action,
		"valid":      boolToUInt8(s.Valid),
		"reason":     s.Reason,
	}, nil
}

// DecodeSale maps a pos.sale payload to a sales_events_raw row. Pure.
func DecodeSale(data []byte) (map[string]any, error) {
	var s events.POSSale
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("chsink: decode pos.sale: %w", err)
	}
	if s.OrgID == uuid.Nil {
		return nil, fmt.Errorf("chsink: pos.sale missing org_id")
	}
	ts := s.OccurredAt
	if ts.IsZero() {
		ts = time.Now()
	}
	return map[string]any{
		"org_id":       s.OrgID.String(),
		"ts":           ts.UTC().Format(chTimeLayout),
		"store_id":     uuid.Nil.String(),
		"source":       s.Source,
		"order_id":     s.OrderID,
		"sku":          s.SKU,
		"qty":          s.Qty,
		"amount_cents": s.AmountCents,
		"currency":     s.Currency,
		"channel":      s.Channel,
	}, nil
}

// Sink writes decoded events into ClickHouse.
type Sink struct {
	ch Inserter
}

// New builds a Sink.
func New(ch Inserter) *Sink {
	return &Sink{ch: ch}
}

// HandleScan decodes and inserts one scan.recorded event.
func (s *Sink) HandleScan(ctx context.Context, data []byte) error {
	row, err := DecodeScan(data)
	if err != nil {
		return err
	}
	return s.ch.Insert(ctx, scanTable, []map[string]any{row})
}

// HandleSale decodes and inserts one pos.sale event.
func (s *Sink) HandleSale(ctx context.Context, data []byte) error {
	row, err := DecodeSale(data)
	if err != nil {
		return err
	}
	return s.ch.Insert(ctx, saleTable, []map[string]any{row})
}
