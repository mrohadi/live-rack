---
description: Work a Notion ticket (LR-{n}) end-to-end — read, branch, TDD, PR
allowed-tools: Read, Write, Edit, Glob, Grep, Bash(git:*), Bash(gh:*), Bash(go:*), Bash(pnpm:*), Bash(make:*), mcp__notion__*
---

# Ticket Workflow

Work on ticket: $ARGUMENTS

## 1. Read Ticket

Fetch from Notion `live-rack · Backlog` where `Ticket ID = $ARGUMENTS`:
- Title, description, acceptance criteria
- Phase, priority, linked phase branch

## 2. Explore Codebase

Use code-review-graph tools FIRST:
- `semantic_search_nodes` for related code
- `get_impact_radius` for blast radius

## 3. Create Branch

```bash
# Derive phase from ticket prefix
# LR-1xx → phase/p1-zones-map
# LR-2xx → phase/p2-scanner  etc.
git checkout phase/p{n}-{slug}
git checkout -b feat/p{n}-$ARGUMENTS-{brief-slug}
```

## 4. TDD Loop

**Red**: Write failing test first
```bash
# Go
go test ./pkg/... -run TestXxx  # must fail

# TypeScript
pnpm -F web test -t "test name"  # must fail
```

**Green**: Minimum code to pass
```bash
make test  # all tests green
```

**Refactor**: Clean up, tests still green

## 5. Commit

```bash
git add {specific files}
git commit -m "feat(p{n}): {what} (passes {test name})

Refs $ARGUMENTS"
```

## 6. PR

```bash
gh pr create \
  --base phase/p{n}-{slug} \
  --title "feat(p{n}): {title}" \
  --body "## Summary
- {bullet}

## Test evidence
- [ ] Unit tests: $(go test ./... 2>&1 | tail -3)
- [ ] TypeScript: $(pnpm typecheck 2>&1 | tail -3)

## Ticket
Refs $ARGUMENTS

🤖 Generated with Claude Code"
```

## 7. Update Notion

Set ticket status → `In review`, add PR URL.
