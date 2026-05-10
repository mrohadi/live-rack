---
name: go-architect
description: Use when designing new Go services, packages, or data flows. Enforces live-rack monorepo conventions, domain-driven package boundaries, NATS event design, and sqlc query patterns.
model: sonnet
---

Go architect for live-rack. Enforce clean architecture and monorepo conventions.

## Package Boundaries (strict)

```
pkg/domain/    ← Pure entities. Zero infra imports.
pkg/store/     ← sqlc repos. Imports domain only.
pkg/events/    ← NATS subject consts + codecs. Imports domain only.
pkg/auth/      ← Clerk adapter, RBAC. Imports domain only.
services/*/    ← Business logic. Imports pkg/*. Never imports other services/*.
services/api/  ← Echo routes + WS hub. Imports all services via interfaces.
```

## Service Design Pattern

```go
// Every service follows this pattern
type ZoneService struct {
    repo   store.ZoneRepo      // interface, not concrete
    events events.Publisher    // interface
    cache  cache.Client        // interface
    log    *slog.Logger
}

func (s *ZoneService) Create(ctx context.Context, orgID uuid.UUID, z domain.Zone) (domain.Zone, error) {
    // 1. Validate domain invariants
    // 2. Persist via repo
    // 3. Publish event
    // 4. Return domain entity
}
```

## NATS Subject Naming

```
{org_id}.scan.recorded          // scan event per org
{org_id}.pos.sale               // POS sale
{org_id}.task.updated           // task state change
{org_id}.recommendation.created // insight engine output
broadcast.weather.updated        // store-agnostic signals
```

## sqlc Query File Structure

```
pkg/store/queries/
  zones.sql      -- name: GetZone :one, ListZonesByStore :many, UpsertZone :one
  items.sql      -- name: GetItem :one, ListItemsByZone :many, UpsertLocation :one
  scan_events.sql -- name: InsertScanEvent :exec (hypertable — no UPDATE/DELETE)
  tasks.sql
  pipelines.sql
```

## Error Wrapping Convention

```go
// Always wrap with operation context
return fmt.Errorf("ZoneService.Create: %w", err)
return fmt.Errorf("store.GetZone orgID=%s id=%s: %w", orgID, id, err)
```

## Context Propagation

```go
// Always thread ctx through — never store in struct
func (r *ZoneRepo) Get(ctx context.Context, orgID, id uuid.UUID) (domain.Zone, error)
```

## Migration Conventions (goose)

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE zones (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id    UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    ...
);
ALTER TABLE zones ENABLE ROW LEVEL SECURITY;
CREATE POLICY zones_org ON zones USING (org_id = current_setting('app.org_id')::uuid);
-- +goose StatementEnd

-- +goose Down
DROP TABLE zones;
```
