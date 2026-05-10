#!/usr/bin/env bash
# Run golangci-lint per Go module for a set of staged files.
# Usage: golangci-lint-staged.sh <file1> <file2> ...
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
CONFIG="$REPO_ROOT/.golangci.yml"

find_module_root() {
  local dir
  dir="$(cd "$(dirname "$1")" && pwd)"
  while [ "$dir" != "$REPO_ROOT" ] && [ "$dir" != "/" ]; do
    if [ -f "$dir/go.mod" ]; then
      echo "$dir"
      return
    fi
    dir="$(dirname "$dir")"
  done
  if [ -f "$REPO_ROOT/go.mod" ]; then
    echo "$REPO_ROOT"
  fi
}

mod_roots=""
for file in "$@"; do
  mod_root="$(find_module_root "$file")"
  if [ -n "$mod_root" ]; then
    mod_roots="$mod_roots
$mod_root"
  fi
done

echo "$mod_roots" | sort -u | grep -v '^$' | while IFS= read -r mod_root; do
  echo "golangci-lint: $mod_root"
  (cd "$mod_root" && golangci-lint run --config "$CONFIG" --fix=false ./...)
done
