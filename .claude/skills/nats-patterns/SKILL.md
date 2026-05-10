---
name: nats-patterns
description: NATS JetStream event patterns for live-rack. Use when publishing scan/sale/recommendation events, building the WS fan-out hub, or writing the ingest worker. Covers subject naming, codec, consumer setup.
---

# NATS Patterns — live-rack

## Subject Naming

```
{org_id}.scan.recorded          # item scanned + validated
{org_id}.pos.sale               # POS sale from Shopify/Square
{org_id}.zone.updated           # zone CRUD
{org_id}.task.updated           # task state change
{org_id}.recommendation.created # insight engine output
broadcast.weather.updated        # org-agnostic weather signal
```

## Event Struct Convention

```go
// pkg/events/scan.go
package events

type ScanRecorded struct {
    OrgID     uuid.UUID `json:"org_id"`
    ScannerID string    `json:"scanner_id"`
    SKU       string    `json:"sku"`
    ZoneID    string    `json:"zone_id"`
    Valid      bool     `json:"valid"`
    Reason    string    `json:"reason,omitempty"`
    TS        time.Time `json:"ts"`
}

const SubjectScanRecorded = "%s.scan.recorded"

func ScanSubject(orgID uuid.UUID) string {
    return fmt.Sprintf(SubjectScanRecorded, orgID)
}
```

## Publisher Interface

```go
// pkg/events/publisher.go
type Publisher interface {
    Publish(ctx context.Context, subject string, v any) error
}

type natsPublisher struct {
    js nats.JetStreamContext
}

func (p *natsPublisher) Publish(ctx context.Context, subject string, v any) error {
    b, err := json.Marshal(v)
    if err != nil {
        return fmt.Errorf("events.Publish marshal: %w", err)
    }
    _, err = p.js.PublishAsync(subject, b)
    return err
}
```

## JetStream Stream Setup (main.go)

```go
js, _ := nc.JetStream()
_, err := js.AddStream(&nats.StreamConfig{
    Name:      "LIVE_RACK",
    Subjects:  []string{"*.scan.recorded", "*.pos.sale", "*.zone.updated", "*.task.updated", "*.recommendation.created", "broadcast.*"},
    MaxAge:    24 * time.Hour,
    Storage:   nats.FileStorage,
    Retention: nats.LimitsPolicy,
})
```

## WS Hub Fan-out

```go
// services/api/internal/ws/hub.go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]map[*Client]struct{} // orgID → clients
    sub     *nats.Subscription
}

func (h *Hub) Subscribe(js nats.JetStreamContext) error {
    _, err := js.Subscribe("*.>", func(msg *nats.Msg) {
        orgID := extractOrgID(msg.Subject)
        h.mu.RLock()
        for c := range h.clients[orgID] {
            c.send <- msg.Data
        }
        h.mu.RUnlock()
        _ = msg.Ack()
    }, nats.DeliverLastPerSubjectPolicy())
    return err
}
```

## Ingest Worker (NATS → ClickHouse)

```go
// services/ingest/worker.go
func (w *Worker) Run(ctx context.Context) error {
    _, err := w.js.Subscribe("*.scan.recorded", func(msg *nats.Msg) {
        var ev events.ScanRecorded
        if err := json.Unmarshal(msg.Data, &ev); err != nil {
            w.log.Error("bad scan event", "err", err)
            _ = msg.Nak()
            return
        }
        if err := w.ch.InsertScanEvent(ctx, ev); err != nil {
            w.log.Error("clickhouse insert", "err", err)
            _ = msg.NakWithDelay(5 * time.Second)
            return
        }
        _ = msg.Ack()
    }, nats.Durable("ingest-scan"), nats.AckExplicit())
    return err
}
```
