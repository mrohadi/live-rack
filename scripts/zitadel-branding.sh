#!/usr/bin/env bash
# Apply live-rack branding to Zitadel's hosted login (LabelPolicy), so the
# sign-in/register pages match our palette (primary #2563eb, Inter).
#
# Usage:
#   OIDC_ISSUER=http://localhost:8081 \
#   ZITADEL_MGMT_TOKEN=<service-account PAT> \
#   ./scripts/zitadel-branding.sh
#
# The service account needs the IAM/org policy-write grant. Logo + font are
# uploaded separately (multipart) via the console; this script sets colors +
# theme flags and activates the policy.
set -euo pipefail

ISSUER="${OIDC_ISSUER:?set OIDC_ISSUER}"
TOKEN="${ZITADEL_MGMT_TOKEN:?set ZITADEL_MGMT_TOKEN}"
BASE="${ISSUER%/}/management/v1"

auth=(-H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json")

echo "→ updating label policy"
curl -fsS -X PUT "${BASE}/policies/label" "${auth[@]}" -d '{
  "primaryColor":        "#2563eb",
  "backgroundColor":     "#f8fafc",
  "warnColor":           "#dc2626",
  "fontColor":           "#0f172a",
  "primaryColorDark":    "#3b82f6",
  "backgroundColorDark": "#0f172a",
  "warnColorDark":       "#ef4444",
  "fontColorDark":       "#f8fafc",
  "hideLoginNameSuffix": true,
  "disableWatermark":    false
}' >/dev/null

echo "→ activating label policy"
curl -fsS -X POST "${BASE}/policies/label/_activate" "${auth[@]}" -d '{}' >/dev/null

echo "✓ branding applied. Hard-refresh the login page to see it."
