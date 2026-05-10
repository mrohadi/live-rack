---
name: pre-commit-checks
description: Run lint, format, type-check, and tests before any git commit or gh pr create. ALWAYS invoke before committing or opening a PR.
---

# Pre-Commit Checks — live-rack

Run all quality gates before `git commit` or `gh pr create`. All steps must exit 0.

## Steps

### 1. Go lint + format (if any `.go` files staged)
```bash
make lint
```
If goimports violations: fix import grouping (stdlib / third-party / local, blank line between groups), then re-stage.

### 2. Prettier check (if any `.ts`, `.tsx`, `.css` files staged)
```bash
make prettier-check
```
On failure, auto-fix and re-stage:
```bash
make prettier-fix
git add <fixed-files>
```

### 3. TypeScript type-check (if any `.ts`/`.tsx` files staged)
```bash
make typecheck
```

### 4. Tests (always)
```bash
make test
```

## Rules

- Never use `--no-verify` unless user explicitly requests it.
- Never commit if any step fails — fix first.
- On PR creation, note which checks passed in PR description.
- `make lint` covers: golangci-lint (Go) + eslint (TS/TSX).
- `make prettier-check` covers: `.ts`, `.tsx`, `.css` under `apps/web/src/`.
