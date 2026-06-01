#!/usr/bin/env bash
# Apply live-rack branding to Zitadel's hosted login so the sign-in / register
# pages match our palette (primary #2563eb, Inter) and default to English.
#
# Mint a service-account PAT once (see scripts/README or the PR notes):
#   Console → Users → Service Users → New → generate Personal Access Token,
#   then grant it Manager role "IAM_OWNER" (Instance) so it can write policies.
#
# Usage:
#   OIDC_ISSUER=http://localhost:8081 \
#   ZITADEL_MGMT_TOKEN=<service-account PAT> \
#   ./scripts/zitadel-branding.sh
#
# Logo + custom font are multipart uploads — do those in the console; this
# script sets colors, theme flags, language, and activates the policy.
set -euo pipefail

ISSUER="${OIDC_ISSUER:?set OIDC_ISSUER}"
TOKEN="${ZITADEL_MGMT_TOKEN:?set ZITADEL_MGMT_TOKEN}"
MGMT="${ISSUER%/}/management/v1"
ADMIN="${ISSUER%/}/admin/v1"

auth=(-H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json")

echo "→ updating label policy (colors + theme)"
curl -fsS -X PUT "${MGMT}/policies/label" "${auth[@]}" -d '{
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
curl -fsS -X POST "${MGMT}/policies/label/_activate" "${auth[@]}" -d '{}' >/dev/null

# Force English on the hosted login regardless of the browser Accept-Language.
echo "→ restricting instance language to English"
curl -fsS -X PUT "${ADMIN}/restrictions" "${auth[@]}" \
  -d '{"allowedLanguages":["en"]}' >/dev/null || \
  echo "  (restrictions endpoint unavailable on this version — skipped)"

echo "✓ branding applied. Hard-refresh the login page (incognito) to see it."
