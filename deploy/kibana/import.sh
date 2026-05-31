#!/usr/bin/env bash
# Import live-rack Kibana saved objects (dashboards, visualizations, index
# patterns) from version-controlled NDJSON. Idempotent: overwrites by id.
#
#   KIBANA_URL=http://localhost:5601 KIBANA_AUTH=elastic:changeme ./deploy/kibana/import.sh
set -euo pipefail

KIBANA_URL="${KIBANA_URL:-http://localhost:5601}"
KIBANA_AUTH="${KIBANA_AUTH:-elastic:changeme}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Importing saved objects into ${KIBANA_URL} ..."
curl -sS -f -u "${KIBANA_AUTH}" \
  -X POST "${KIBANA_URL}/api/saved_objects/_import?overwrite=true" \
  -H "kbn-xsrf: true" \
  --form file=@"${DIR}/saved-objects.ndjson" \
  | tee /dev/stderr | grep -q '"success":true'

echo "Kibana saved objects imported."
