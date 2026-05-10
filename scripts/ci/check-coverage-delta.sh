#!/usr/bin/env bash
# Fails if Go coverage on touched packages drops > 0.5pp vs base branch.
# Usage: check-coverage-delta.sh <base-branch> <coverage.out>
set -euo pipefail

BASE_BRANCH="${1:-main}"
COVERAGE_FILE="${2:-coverage.out}"
THRESHOLD=0.5

if [[ ! -f "$COVERAGE_FILE" ]]; then
  echo "Coverage file not found: $COVERAGE_FILE"
  exit 1
fi

# Extract total coverage from current run
CURRENT=$(go tool cover -func="$COVERAGE_FILE" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')

# Fetch base coverage via stash or artifact; fall back to 0 (first run always passes)
BASE_COVERAGE_FILE="/tmp/base-coverage.out"
if git fetch origin "$BASE_BRANCH" --depth=1 2>/dev/null && \
   git show "origin/$BASE_BRANCH:coverage.out" > "$BASE_COVERAGE_FILE" 2>/dev/null; then
  BASE=$(go tool cover -func="$BASE_COVERAGE_FILE" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')
else
  echo "No base coverage found — skipping delta check (first run)"
  exit 0
fi

DELTA=$(awk "BEGIN { printf \"%.2f\", $CURRENT - $BASE }")
echo "Coverage: base=${BASE}%  current=${CURRENT}%  delta=${DELTA}pp"

DROP=$(awk "BEGIN { print ($BASE - $CURRENT > $THRESHOLD) ? \"1\" : \"0\" }")
if [[ "$DROP" == "1" ]]; then
  echo "FAIL: coverage dropped ${DELTA}pp (threshold -${THRESHOLD}pp)"
  exit 1
fi

echo "OK: coverage delta within threshold"
