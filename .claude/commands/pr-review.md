---
description: Review a PR using the code-reviewer agent. Pass PR number or URL.
allowed-tools: Read, Glob, Grep, Bash(git:*), Bash(gh:*), Bash(go:*), Bash(pnpm:*)
---

# PR Review

PR: $ARGUMENTS

## Steps

1. Checkout PR: `gh pr checkout $ARGUMENTS`
2. Get diff: `git diff origin/$(gh pr view $ARGUMENTS --json baseRefName -q .baseRefName)...HEAD`
3. Apply code-reviewer.md checklist to every modified file
4. Comment findings on PR:

```bash
gh pr comment $ARGUMENTS --body "## Code Review

### Critical
{issues or 'None'}

### Warnings  
{issues or 'None'}

### Suggestions
{issues or 'None'}

### Test coverage
- Go coverage delta: $(go test ./... -cover 2>&1 | grep -E 'coverage|FAIL' | head -10)
- TypeScript: $(pnpm typecheck 2>&1 | tail -5)

🤖 Claude Code review"
```

5. Approve or request changes via `gh pr review $ARGUMENTS --approve` / `--request-changes`
