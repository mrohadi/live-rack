---
name: go-patterns
description: Go coding patterns for live-rack. Use when writing Go services, domain entities, sqlc repos, Echo handlers, or NATS publishers. Covers error wrapping, context propagation, dependency injection, and interface design.
---

# Go Patterns — live-rack

## Domain Entity Pattern

```go
// pkg/domain/zone.go
package domain

import "github.com/google/uuid"

type ZoneType string

const (
    ZoneTypeReceiving  ZoneType = "receiving"
    ZoneTypeApparel    ZoneType = "apparel"
    ZoneTypeElectronic ZoneType = "electronic"
    ZoneTypeCold       ZoneType = "cold"
    ZoneTypeReturns    ZoneType = "returns"
    ZoneTypeOutbound   ZoneType = "outbound"
    ZoneTypeBulk       ZoneType = "bulk"
    ZoneTypeShowroom   ZoneType = "showroom"
    ZoneTypeStaging    ZoneType = "staging"
    ZoneTypeHome       ZoneType = "home"
)

type Zone struct {
    ID         uuid.UUID
    OrgID      uuid.UUID
    StoreID    uuid.UUID
    Name       string
    Type       ZoneType
    X, Y, W, H int
    Capacity   int
    Color      string
}

func (z Zone) OccupancyPct(current int) float64 {
    if z.Capacity == 0 {
        return 0
    }
    return float64(current) / float64(z.Capacity) * 100
}
```

## Service Pattern

```go
// services/api/internal/zones/service.go
package zones

type Service struct {
    repo   ZoneRepo
    events EventPublisher
    log    *slog.Logger
}

type ZoneRepo interface {
    Get(ctx context.Context, orgID, id uuid.UUID) (domain.Zone, error)
    List(ctx context.Context, orgID, storeID uuid.UUID) ([]domain.Zone, error)
    Upsert(ctx context.Context, z domain.Zone) (domain.Zone, error)
    Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type EventPublisher interface {
    Publish(ctx context.Context, subject string, v any) error
}

func NewService(repo ZoneRepo, events EventPublisher, log *slog.Logger) *Service {
    return &Service{repo: repo, events: events, log: log}
}

func (s *Service) Upsert(ctx context.Context, orgID uuid.UUID, z domain.Zone) (domain.Zone, error) {
    z.OrgID = orgID
    result, err := s.repo.Upsert(ctx, z)
    if err != nil {
        return domain.Zone{}, fmt.Errorf("zones.Service.Upsert: %w", err)
    }
    if err := s.events.Publish(ctx, fmt.Sprintf("%s.zone.updated", orgID), result); err != nil {
        s.log.Warn("failed to publish zone.updated", "err", err)
    }
    return result, nil
}
```

## Echo Handler Pattern

```go
// services/api/internal/zones/handler.go
func (h *Handler) UpsertZone(c echo.Context) error {
    orgID := auth.OrgIDFromCtx(c.Request().Context())

    var req UpsertZoneRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
    }

    zone, err := h.svc.Upsert(c.Request().Context(), orgID, req.ToDomain())
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "upsert failed")
    }
    return c.JSON(http.StatusOK, ZoneResponse{}.FromDomain(zone))
}
```

## Table-Driven Test Pattern

```go
func TestZoneService_Upsert(t *testing.T) {
    tests := []struct {
        name    string
        input   domain.Zone
        wantErr bool
    }{
        {"valid zone", domain.Zone{Name: "A1", Type: domain.ZoneTypeReceiving, Capacity: 200}, false},
        {"zero capacity", domain.Zone{Name: "A1", Capacity: 0}, false},
        {"empty name", domain.Zone{}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := newTestService(t)
            _, err := svc.Upsert(context.Background(), uuid.New(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("got err=%v, wantErr=%v", err, tt.wantErr)
            }
        })
    }
}
```
