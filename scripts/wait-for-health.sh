#!/usr/bin/env bash
# Poll a health URL until it returns HTTP 200 or the timeout (seconds) elapses.
#   ./scripts/wait-for-health.sh http://localhost:8080/healthz 120
set -euo pipefail

URL="${1:?usage: wait-for-health.sh <url> [timeout_seconds]}"
TIMEOUT="${2:-120}"
DEADLINE=$(( $(date +%s) + TIMEOUT ))

until curl -fsS -o /dev/null "$URL"; do
  if [ "$(date +%s)" -ge "$DEADLINE" ]; then
    echo "timed out waiting for $URL after ${TIMEOUT}s" >&2
    exit 1
  fi
  sleep 2
done
echo "healthy: $URL"
