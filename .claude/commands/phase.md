---
description: Start or close a project phase (P0–P11). Creates/merges phase branch and updates Notion.
allowed-tools: Bash(git:*), Bash(gh:*), Bash(make:*), mcp__notion__*
---

# Phase Management

Action and phase: $ARGUMENTS
(e.g. "start p1" or "close p3")

## Start Phase

```bash
# Cut phase branch from main
git checkout main && git pull
git checkout -b phase/p{n}-{slug}
git push -u origin phase/p{n}-{slug}
```

Mark all phase tickets in Notion → `Todo`.

Print phase ticket list for the team.

## Close Phase

### Verify exit criteria
- [ ] All tickets in Notion → `Done`
- [ ] All tests green: `make test`
- [ ] E2E smoke: `pnpm -F web exec playwright test --project=smoke`
- [ ] No open PRs targeting phase branch

### Merge + tag
```bash
git checkout main
git merge --no-ff phase/p{n}-{slug} -m "release: merge phase/p{n}-{slug}

$(gh pr list --state merged --base phase/p{n}-{slug} --json title,number | python3 -c 'import sys,json; [print(f"- PR #{i[\"number\"]}: {i[\"title\"]}") for i in json.load(sys.stdin)]')"

git tag v0.{n}.0
git push origin main --tags
git push origin --delete phase/p{n}-{slug}
```

### Update Notion
Mark phase epic → `Done`. Post release notes to team channel.
