# live-rack

SaaS Warehouse Zoning, Real-Time Tracking & Analytics Platform.

---

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.22+ | `brew install go` |
| Node.js | 20+ | `brew install node` |
| pnpm | 9+ | `npm i -g pnpm@9` |
| Docker + Compose | v2 | [docker.com](https://docker.com) |
| golangci-lint | latest | `brew install golangci-lint` |
| gitleaks | latest | `brew install gitleaks` |
| goose | latest | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| sqlc | latest | `brew install sqlc` |

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/your-org/live-rack && cd live-rack

# 2. Install frontend deps
pnpm install

# 3. Copy env files and fill in values
cp apps/web/.env.example apps/web/.env.local

# 4. Start all infrastructure (DB, NATS, Redis, ClickHouse, MinIO, observability)
make dev

# 5. Run migrations
DATABASE_URL=postgres://postgres:postgres@localhost:5432/liverack?sslmode=disable make migrate-up

# 6. Install pre-commit hooks (one-time per machine)
make hooks-install

# 7. Start API
cd services/api && DATABASE_URL=postgres://postgres:postgres@localhost:5432/liverack?sslmode=disable CLERK_SECRET_KEY=sk_test_xxx go run .

# 8. Start web app (separate terminal)
pnpm -F web dev
```

---

## Environment Variables

### `apps/web/.env.local`

```env
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
VITE_API_URL=http://localhost:8080
```

### API service (shell or `.env`)

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/liverack?sslmode=disable
CLERK_SECRET_KEY=sk_test_...
CLERK_WEBHOOK_SECRET=whsec_...          # optional — only for Clerk webhooks
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317  # optional — enable tracing
LOG_LEVEL=info                          # debug | info | warn | error
PORT=8080
```

---

## Make Targets

```bash
make dev              # docker compose up -d (all infra)
make dev-status       # show running containers
make migrate-up       # run goose migrations (needs DATABASE_URL)
make migrate-status   # show migration state
make generate         # sqlc generate
make seed             # load fixture data
make seed-reset       # truncate + reseed
make test             # go test -race ./... + vitest
make lint             # golangci-lint + eslint
make typecheck        # tsc --noEmit
make build            # compile all Go services + frontend
make hooks-install    # install lefthook pre-commit hooks
make notion-seed      # seed Notion backlog (needs NOTION_API_KEY + NOTION_PARENT_PAGE_ID)
make clean            # rm build artefacts + docker compose down -v
```

---

## Service URLs (local)

| Service | URL | Notes |
|---|---|---|
| Web app | http://localhost:3000 | `pnpm -F web dev` |
| API | http://localhost:8080 | `go run ./services/api` |
| API health | http://localhost:8080/healthz | |
| API metrics | http://localhost:8080/metrics | Prometheus scrape |
| Postgres | `localhost:5432` | user: postgres / pw: postgres / db: liverack |
| NATS | `localhost:4222` | monitoring: http://localhost:8222 |
| ClickHouse | http://localhost:8123 | HTTP interface |
| Redis | `localhost:6379` | |
| MinIO console | http://localhost:9001 | user: minioadmin / pw: minioadmin |
| Grafana | http://localhost:3000 | anonymous admin (conflicts with web dev on same port — see note) |
| Prometheus | http://localhost:9090 | |
| Loki | http://localhost:3100 | |
| Tempo | http://localhost:3200 | |
| OTLP gRPC | `localhost:4317` | services → collector |

> **Port conflict**: Grafana and Vite dev server both default to `:3000`. Run `pnpm -F web dev -- --port 5173` or change Grafana's port in `deploy/docker/docker-compose.yml`.

---

## Running Tests

```bash
# All (Go + frontend)
make test

# Go only with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Frontend only
pnpm -F web test

# Frontend with coverage report
pnpm -F web test --run --coverage

# Single Go package
go test -v ./pkg/auth/...
```

---

## Database Migrations

```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/liverack?sslmode=disable

# Apply all pending
goose -dir migrations postgres "$DATABASE_URL" up

# Check status
goose -dir migrations postgres "$DATABASE_URL" status

# Rollback one
goose -dir migrations postgres "$DATABASE_URL" down
```

---

## Notion Backlog Seed

Creates `live-rack · Backlog` database and seeds all 66 tickets (P0–P11).

```bash
# First run — creates DB, prints NOTION_DATABASE_ID
NOTION_API_KEY=secret_xxx NOTION_PARENT_PAGE_ID=<page-id> make notion-seed

# Re-run — skips DB creation, idempotent
NOTION_API_KEY=secret_xxx NOTION_DATABASE_ID=<db-id> make notion-seed
```

---

## Pre-commit Hooks

Powered by [lefthook](https://github.com/evilmartians/lefthook). Install once:

```bash
make hooks-install
# also requires: brew install golangci-lint gitleaks
```

Hooks run on `git commit`:
- `gofmt` — formats staged `.go` files
- `golangci-lint` — lints staged `.go` files
- `eslint` — lints staged `.ts/.tsx`
- `prettier` — checks staged `.ts/.tsx/.css`
- `gitleaks` — blocks secrets from committing

Commit messages must follow Conventional Commits + include `Refs LR-{n}`:
```
feat(p1): add zone drag-resize
<blank line>
Refs LR-104
```

Skip hooks (CI only, not for normal use):
```bash
LEFTHOOK=0 git commit ...
```

---

## Observability

Grafana at http://localhost:3000 (anonymous admin).

Provisioned datasources: Prometheus, Loki, Tempo — linked for trace↔log correlation.

Dashboard: `live-rack · API Overview` — req/s, p95 latency, error rate, traces.

Enable tracing in API:
```bash
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 go run ./services/api
```

---

## Project Structure

```
live-rack/
├── apps/
│   ├── web/                React SPA (Vite + TypeScript + Tailwind)
│   └── scanner/            PWA — camera + WebHID barcode scanning
├── services/
│   ├── api/                Go Echo HTTP + WebSocket gateway
│   ├── ingest/             NATS → ClickHouse worker
│   ├── rollup/             Cron analytics jobs
│   ├── integrations/       Shopify, Square, Stripe, Shippo, Weather, Transit
│   └── insight/            Recommendation engine
├── pkg/
│   ├── domain/             Pure entities — Zone, Item, Task, Pipeline, User, Org
│   ├── store/              sqlc-generated Postgres repos
│   ├── events/             NATS subject schemas
│   ├── auth/               Clerk adapter, RBAC
│   └── observability/      OTel bootstrap
├── deploy/
│   ├── docker/             docker-compose (infra + observability)
│   ├── grafana/            Prometheus, Loki, Tempo, OTel collector configs + dashboards
│   └── terraform/          AWS modules — network, db, cache, bus, observability
├── migrations/             goose SQL files
├── scripts/
│   ├── ci/                 Coverage delta + commit message check scripts
│   └── notion-seed/        Notion backlog seeder (Go)
├── .github/workflows/      CI — lint, test, coverage, security, commit-lint
├── lefthook.yml            Pre-commit hook definitions
└── Makefile                All developer commands
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 18 + Vite + TypeScript + Tailwind + shadcn/ui + Zustand + TanStack Query |
| Backend | Go 1.22 + Echo + sqlc + goose |
| Realtime | NATS JetStream + WebSocket gateway |
| Primary DB | PostgreSQL 16 + TimescaleDB |
| Analytics | ClickHouse |
| Cache | Redis 7 |
| Object storage | MinIO (S3-compatible) |
| Auth | Clerk — multi-tenant, SSO, 2FA |
| Scanner PWA | @zxing/browser + WebHID + IndexedDB offline queue |
| Observability | Grafana + Loki + Tempo + Prometheus + OpenTelemetry |
| IaC | Terraform (AWS) |
| CI/CD | GitHub Actions |

---

## Architecture

```mermaid
flowchart LR
  subgraph Clients
    WEB[Web SPA<br/>React+Tailwind]
    PWA[Scanner PWA<br/>Camera/WebHID]
  end

  subgraph Edge
    GW[API Gateway<br/>Go Echo :8080]
    WS[WebSocket Gateway]
  end

  subgraph Core[Go services]
    AUTH[Auth/Tenant]
    ZONE[Zones & Map]
    INV[Inventory & Scan]
    POS[POS Sync]
    TASK[Tasks]
    PIPE[Pipelines]
    INTG[Integrations Hub]
  end

  subgraph Data
    PG[(Postgres+Timescale<br/>:5432)]
    CH[(ClickHouse<br/>:8123)]
    REDIS[(Redis :6379)]
    S3[(MinIO :9002)]
  end

  subgraph Bus
    NATS[NATS JetStream<br/>:4222]
  end

  subgraph Obs[Observability]
    PROM[Prometheus :9090]
    LOKI[Loki :3100]
    TEMPO[Tempo :3200]
    GRAF[Grafana :3000]
  end

  WEB & PWA --> GW
  WEB & PWA <--> WS
  GW --> AUTH & ZONE & INV & POS & TASK & PIPE & INTG
  WS <--> NATS
  INV & POS --> NATS
  AUTH & ZONE & INV & TASK --> PG
  INV --> NATS --> CH
  INV --> S3
  AUTH --> REDIS
  GW --> PROM
  PROM & LOKI & TEMPO --> GRAF
```
