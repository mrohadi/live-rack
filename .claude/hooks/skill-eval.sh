#!/usr/bin/env bash
# Evaluates user prompt and suggests relevant skill to load.
# Runs on UserPromptSubmit — output shown as pre-prompt context.

PROMPT="${CLAUDE_USER_PROMPT:-}"

hint=""

if echo "$PROMPT" | grep -qiE "go |\.go|service|domain|sqlc|repo|handler|echo|goose|migration"; then
  hint="go-patterns"
fi

if echo "$PROMPT" | grep -qiE "test|spec|tdd|vitest|playwright|testcontainer|mock|factory"; then
  hint="tdd-workflow"
fi

if echo "$PROMPT" | grep -qiE "react|component|tsx|tailwind|zustand|tanstack|query|hook|page"; then
  hint="react-patterns"
fi

if echo "$PROMPT" | grep -qiE "nats|jetstream|websocket|ws |event|publish|subscribe|ingest|fan.?out"; then
  hint="nats-patterns"
fi

if echo "$PROMPT" | grep -qiE "scanner|barcode|zxing|zebra|webhid|offline|indexeddb|scan"; then
  hint="scanner-patterns"
fi

if [ -n "$hint" ]; then
  echo "{\"feedback\": \"Skill hint: consider loading '.claude/skills/${hint}/SKILL.md' for this task.\"}"
fi

exit 0
