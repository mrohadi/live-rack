---
name: tdd-workflow
description: TDD red-green-refactor workflow for live-rack. Use when writing any new function, component, or endpoint. Covers Go unit/repo/service tests and TypeScript Vitest + Testing Library patterns.
---

# TDD Workflow — live-rack

## The Loop (always)

1. **Red**: Write failing test. Commit: `test(p{n}): add failing test for {X}`
2. **Green**: Minimum code to pass. Commit: `feat(p{n}): {X} (passes {test name})`
3. **Refactor**: Clean internals. Tests still green. Commit: `refactor(p{n}): clean up {X}`

## Go — Unit Test (pure domain)

```go
// pkg/domain/zone_test.go
func TestZone_OccupancyPct(t *testing.T) {
    z := domain.Zone{Capacity: 200}
    assert.Equal(t, 50.0, z.OccupancyPct(100))
    assert.Equal(t, 0.0, z.OccupancyPct(0))
    assert.Equal(t, 0.0, domain.Zone{Capacity: 0}.OccupancyPct(10)) // no divide-by-zero
}
```

## Go — Repo Test (real Postgres via testcontainers)

```go
// pkg/store/zones_test.go
func TestZoneRepo_Upsert(t *testing.T) {
    db := testdb.New(t) // starts postgres container, runs migrations
    repo := store.NewZoneRepo(db)

    zone := domain.Zone{
        OrgID:    testOrg,
        Name:     "A1 · Receiving",
        Type:     domain.ZoneTypeReceiving,
        Capacity: 200,
    }
    got, err := repo.Upsert(context.Background(), zone)
    require.NoError(t, err)
    assert.NotEmpty(t, got.ID)
    assert.Equal(t, zone.Name, got.Name)
}
```

## Go — testdb helper

```go
// pkg/testutil/testdb.go
func New(t *testing.T) *sql.DB {
    t.Helper()
    ctx := context.Background()
    pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "timescale/timescaledb:latest-pg16",
            ExposedPorts: []string{"5432/tcp"},
            WaitingFor:   wait.ForListeningPort("5432/tcp"),
        },
        Started: true,
    })
    require.NoError(t, err)
    t.Cleanup(func() { _ = pg.Terminate(ctx) })

    host, _ := pg.Host(ctx)
    port, _ := pg.MappedPort(ctx, "5432")
    dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", host, port.Port())

    db, err := sql.Open("pgx", dsn)
    require.NoError(t, err)
    runMigrations(t, dsn)
    return db
}
```

## TypeScript — Component Test

```typescript
// apps/web/src/features/map/__tests__/ZoneCard.test.tsx
import { render, screen } from '@testing-library/react'
import { ZoneCard } from '../ZoneCard'

const mockZone = (overrides = {}) => ({
  id: 'A1',
  name: 'A1 · Receiving',
  type: 'receiving' as const,
  items: 142,
  capacity: 200,
  color: '#7c3aed',
  ...overrides,
})

test('shows occupancy percentage', () => {
  render(<ZoneCard zone={mockZone()} />)
  expect(screen.getByText('71%')).toBeInTheDocument()
})

test('shows warning when capacity > 90%', () => {
  render(<ZoneCard zone={mockZone({ items: 185, capacity: 200 })} />)
  expect(screen.getByRole('alert')).toHaveTextContent('Near capacity')
})
```

## TypeScript — API Mock with MSW

```typescript
// apps/web/src/mocks/handlers.ts
import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v1/zones', () => {
    return HttpResponse.json({ zones: [mockZone()] })
  }),
  http.post('/api/v1/scan', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json({ valid: true, zone: 'A2' })
  }),
]
```

## Coverage Targets

| Layer | Target |
|---|---|
| `pkg/domain/` | ≥85% |
| `pkg/store/` | ≥80% |
| Services | ≥75% |
| React features | ≥70% |
| E2E (Playwright) | 1 happy-path + edge cases per phase |
