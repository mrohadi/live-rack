# live-rack · Claude Code Configuration

## Stack

| Layer | Technology |
|---|---|
| Frontend | React 18 + Vite + TypeScript + Tailwind + shadcn/ui + Zustand + TanStack Query |
| Backend | Go 1.22 + Echo + sqlc + goose |
| Realtime | NATS JetStream + WebSocket gateway |
| Primary DB | PostgreSQL 16 + TimescaleDB |
| Analytics | ClickHouse |
| Cache | Redis 7 |
| Auth | Clerk (multi-tenant, SSO, 2FA) |
| Scanner PWA | @zxing/browser + WebHID + IndexedDB offline queue |
| Observability | ELK Stack (Elasticsearch + Logstash + Kibana) + Filebeat + APM Server + OpenTelemetry |

## Commands

```bash
# Dev
make dev          # docker compose up (PG+TS, NATS, ClickHouse, Redis, MinIO)
make seed         # Load fixtures from references/live-rack/project/data.jsx
make test         # Run all tests (Go + Vitest)
make lint         # golangci-lint + eslint + prettier
make typecheck    # tsc --noEmit

# Frontend
pnpm -F web dev
pnpm -F web test
pnpm -F web build

# Backend (from services/api/)
go run .
go test ./...
go generate ./...   # sqlc generate

# DB
goose -dir migrations postgres "$DATABASE_URL" up
goose -dir migrations postgres "$DATABASE_URL" status
```

## Key Directories

```
apps/
  web/src/features/       # Feature modules (map, scanner, inventory, tasks, pipelines, analytics, integrations, users)
  web/src/components/     # Shared UI components
  web/src/lib/            # API client, WS client, utils
services/
  api/                    # Echo HTTP + WebSocket gateway
  ingest/                 # NATS → ClickHouse worker
  rollup/                 # Cron analytics jobs
  integrations/           # Shopify, Square, Stripe, Shippo, Weather, Transit adapters
  insight/                # Recommendation engine
pkg/
  domain/                 # Pure entities: Zone, Item, Task, Pipeline, User, Org
  store/                  # sqlc-generated Postgres repos
  events/                 # NATS subject schemas
  auth/                   # Clerk adapter, RBAC
  observability/          # OTel bootstrap
migrations/               # goose SQL files
references/               # Claude Design bundle (read-only reference)
```

## Code Style — Go

- `golangci-lint` config in `.golangci.yml`
- No naked returns
- Context always first arg: `func Foo(ctx context.Context, ...)`
- Errors wrapped with `fmt.Errorf("op: %w", err)` 
- No `interface{}` — use typed interfaces
- `pkg/domain/` entities are pure structs, zero imports from infra

## Code Style — TypeScript / React

- Strict mode — no `any`, use `unknown`
- `interface` over `type` (except unions)
- Early returns, no deep nesting
- Components in `features/` own their data fetching (TanStack Query)
- Shared UI only in `components/`; never call API from there
- Tailwind utility-first; no inline styles

## Git Conventions

- Branch: `phase/p{n}-{slug}` (phase integration), `feat/p{n}-{ticket}-{slug}` (ticket work), `fix/{ticket}-{slug}` (hotfix)
- Commits: Conventional Commits (`feat:`, `fix:`, `chore:`, `test:`, `docs:`, `refactor:`, `perf:`, `build:`, `ci:`)
- Body must include: `Refs LR-{n}`
- No direct push to `main` — always PR with ≥1 review + CI green

## Workflow Rules (review-first)

- **NEVER auto-commit or open a PR after implementing a feature.** Stop after code + tests pass locally and wait for user review.
- After implementation: report what changed (file list + diff summary) and stop. Let user read the diff before any `git add` / `git commit` / `gh pr create`.
- User says "commit" or "open PR" → only then run those commands.
- Same rule for `git push`. Wait for explicit instruction.
- Applies to every ticket, every refactor, every fix.

## TDD Rules

- Write failing test FIRST (`test(p{n}): add failing test for X`)
- Minimum code to pass (`feat(p{n}): X (passes Y test)`)
- Refactor after green (`refactor(p{n}): clean up X`)
- Go: `testcontainers-go` for real Postgres/NATS in repo/service tests
- Frontend: Vitest + Testing Library + MSW for component tests
- E2E: Playwright (add per phase, minimum 1 happy-path + 1 mis-scan)

## Multi-tenancy Rules

- Every table has `org_id UUID NOT NULL`
- Postgres RLS policy per table — never bypass with `SET LOCAL row_security = off`
- All repo functions receive `orgID uuid.UUID` — never derive from context alone
- Clerk org → `org_id` mapping enforced at API gateway middleware

## Security

- Secrets via env — never hardcode, never commit `.env`
- `gitleaks` pre-commit blocks secret commits
- SQL via sqlc only — no raw string interpolation
- HTML output via React — no `dangerouslySetInnerHTML`
- Rate-limit at gateway: 100 req/s per org, 20 scan events/s per scanner

## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

## Skill Activation

Before implementing ANY task, check if a skill applies:

- Writing Go service/domain code → `go-patterns` skill
- Writing tests (Go or TS) → `tdd-workflow` skill  
- Building React components → `react-patterns` skill
- Working on NATS events / WS → `nats-patterns` skill
- Scanner PWA / barcode / WebHID → `scanner-patterns` skill
- **Before any `git commit` or `gh pr create`** → `pre-commit-checks` skill (runs lint, prettier, typecheck, tests)

## Notion Backlog

Tickets tracked in Notion DB `live-rack · Backlog`. Format: `LR-{n}`.
Every commit body must reference the ticket: `Refs LR-{n}`.
Phase cut line: LR-001 through LR-606 = MVP (P0–P6, ~18 weeks).

<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
|------|----------|
| `detect_changes` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context` | Need source snippets for review — token-efficient |
| `get_impact_radius` | Understanding blast radius of a change |
| `get_affected_flows` | Finding which execution paths are impacted |
| `query_graph` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `get_architecture_overview` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes` for code review.
3. Use `get_affected_flows` to understand impact.
4. Use `query_graph` pattern="tests_for" to check coverage.
