#!/usr/bin/env bash
# Validates commit message against Conventional Commits + Refs LR-{n} requirement.
# Usage: check-commit-msg.sh <commit-msg-file>
set -euo pipefail

MSG_FILE="${1:-}"
if [[ -z "$MSG_FILE" || ! -f "$MSG_FILE" ]]; then
  echo "ERROR: commit message file not provided or missing"
  exit 1
fi

MSG=$(cat "$MSG_FILE")

# Strip comments (lines starting with #)
STRIPPED=$(echo "$MSG" | grep -v '^#' | sed '/^$/d')

FIRST_LINE=$(echo "$STRIPPED" | head -1)

# Conventional Commits pattern:
#   type(optional-scope): description
TYPES="feat|fix|chore|test|docs|refactor|perf|build|ci|release"
CC_PATTERN="^($TYPES)(\([a-z0-9/_-]+\))?: .{1,100}$"

if ! echo "$FIRST_LINE" | grep -qE "$CC_PATTERN"; then
  echo ""
  echo "FAIL: invalid commit message format."
  echo ""
  echo "  Got: $FIRST_LINE"
  echo ""
  echo "  Expected: <type>(scope): <description>"
  echo "  Types: feat fix chore test docs refactor perf build ci release"
  echo ""
  exit 1
fi

# Merge commits and release commits skip Refs check
if echo "$FIRST_LINE" | grep -qE "^(release|Merge):"; then
  exit 0
fi

# Body must contain Refs LR-{n} (ticket reference)
if ! echo "$STRIPPED" | grep -qE "Refs LR-[0-9A-Z]+"; then
  echo ""
  echo "FAIL: commit body missing ticket reference."
  echo "  Add: Refs LR-{n}"
  echo ""
  exit 1
fi

exit 0
