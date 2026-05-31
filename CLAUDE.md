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
| Auth | Zitadel (self-hosted OIDC, multi-tenant, SSO, 2FA) — replaced Clerk (LR-005a) |
| Scanner PWA | @zxing/browser + WebHID + IndexedDB offline queue |
| Observability | ELK Stack (Elasticsearch + Logstash + Kibana) + Filebeat + APM Server + OpenTelemetry |

## Commands

```bash
# Dev
make dev          # docker compose up (PG+TS, NATS, ClickHouse, Redis, MinIO)
make seed         # Load fixtures from ui-references/live-rack/data.jsx
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
  auth/                   # Zitadel OIDC verifier (JWKS), JIT provisioning, RBAC
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

## Workflow Rules (guide-first — user writes the code)

User is learning Go + this codebase by writing code themselves. Claude **teaches and guides**, does not implement.

### What Claude does

For every ticket:
1. Plan ordered steps
2. Per step provide: exact bash commands, file paths (absolute), code as fenced blocks the user copies, plain-English explanation, learning URL when a new concept is introduced
3. **Verify + test section** at the end — start command, manual smoke (curl/UI), expected output, automated test command, expected pass output
4. **Suggest** commit message + PR body in fenced blocks; user runs `git add . && git commit / gh pr create` themselves
5. After user reports it merged: refresh `code-review-graph` via the MCP tool

### What Claude does NOT do

- **No `Edit` or `Write` tool on source files** under `apps/`, `services/`, `pkg/`, `migrations/`, etc. User writes them.
- **No `git commit`, `git push`, `gh pr create`** — user runs these
- **No `make`, `go test`, `pnpm test`** to "verify" — user runs, reports back
- Exceptions: meta-files like `.claude/`, `~/.claude/`, plan files, memory, skills, READMEs about workflow. Config / infra (docker-compose, .yml) is OK when user explicitly says "fix it".

### Each new chat = one feature

User opens a new chat per ticket. This base chat holds the workflow rules and overall plan. New chats inherit via this file + memory + plan.

First action in a new chat:
1. Read this file
2. Read `~/.claude/projects/-Users-mrohadi-Projects-live-rack/memory/MEMORY.md` and the feedback files it links
3. Read the ticket from Notion (LR-{n})
4. Read the plan: `~/.claude/plans/you-are-senior-software-lexical-jellyfish.md`
5. Build/refresh code-review-graph if not current
6. Walk the user through, wait for confirmation between phases

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
- Zitadel org → `org_id` mapping enforced at API gateway middleware (JIT-provisioned on first login)

## Auth — Zitadel (self-hosted OIDC)

- Dev: `docker compose -f deploy/docker/docker-compose.yml up -d zitadel zitadel-db`
- Pinned `ghcr.io/zitadel/zitadel:v2.71.9` (legacy hosted login; `OIDC_DEFAULTLOGINURLV2=""`)
- Console: http://localhost:8081/ui/console — admin `admin@localhost` / `Admin123!`
- Project `live-rack` (ID `375294302823120905`); SPA app `web` PKCE, redirect `http://localhost:5173/callback`
- App Token Settings: **JWT** + "Add user roles to access token" + "User roles inside ID token" (required)
- Roles: admin, manager, staff, readonly, service (assign via project Authorizations)
- Org id derived from project-roles claim inner key (`urn:zitadel:iam:org:project:{id}:roles`) — no resourceowner claim
- Env — api: `OIDC_ISSUER`, `OIDC_PROJECT_ID`; web: `VITE_OIDC_ISSUER`, `VITE_OIDC_CLIENT_ID`, `VITE_OIDC_REDIRECT_URI`
- Frontend: `react-oidc-context` + `oidc-client-ts`, `loadUserInfo: true`

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
