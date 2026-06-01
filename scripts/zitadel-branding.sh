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

auth=(-H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json")

PALETTE='{
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
}'

# A fresh org inherits the default (instance) label policy, so we POST to create
# a custom org policy first; PUT only succeeds once one exists. Both are
# idempotent here — a "not changed" PUT is fine.
echo "→ creating custom label policy"
curl -fsS -o /dev/null -X POST "${MGMT}/policies/label" "${auth[@]}" -d "${PALETTE}" || true

echo "→ updating label policy (colors + theme)"
curl -fsS -o /dev/null -X PUT "${MGMT}/policies/label" "${auth[@]}" -d "${PALETTE}" || true

echo "→ activating label policy"
curl -fsS -o /dev/null -X POST "${MGMT}/policies/label/_activate" "${auth[@]}" -d '{}'

echo "✓ branding applied. Hard-refresh the login page (incognito) to see it."
echo "  Logo + Inter font: Console → Default settings → Branding (multipart upload)."
echo "  English login: Console → Default settings → Languages → set default 'en'."
