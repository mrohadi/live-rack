#!/usr/bin/env bash
# Configure Zitadel to send invite / verification emails via Resend's SMTP relay.
# Keeps the native Zitadel invite flow (codes, templates) — no app code involved.
#
# Resend SMTP credentials (https://resend.com/settings):
#   host: smtp.resend.com:587   user: resend   password: <RESEND_API_KEY>
# Sender must be a verified domain/address in your Resend account.
#
# The SMTP config is instance-scoped, so the token MUST be an IAM_OWNER PAT
# (the branding/onboarding service user works only if granted IAM_OWNER).
#
# Usage:
#   OIDC_ISSUER=http://localhost:8081 \
#   ZITADEL_ADMIN_TOKEN=<IAM_OWNER PAT> \
#   RESEND_API_KEY=re_xxx \
#   SMTP_SENDER_ADDRESS=no-reply@yourdomain.com \
#   SMTP_SENDER_NAME="live-rack" \
#   ./scripts/zitadel-smtp.sh
set -euo pipefail

ISSUER="${OIDC_ISSUER:?set OIDC_ISSUER}"
TOKEN="${ZITADEL_ADMIN_TOKEN:?set ZITADEL_ADMIN_TOKEN (IAM_OWNER PAT)}"
KEY="${RESEND_API_KEY:?set RESEND_API_KEY}"
FROM="${SMTP_SENDER_ADDRESS:?set SMTP_SENDER_ADDRESS (verified in Resend)}"
NAME="${SMTP_SENDER_NAME:-live-rack}"
ADMIN="${ISSUER%/}/admin/v1"

auth=(-H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json")

echo "→ adding Resend SMTP provider"
resp=$(curl -fsS -X POST "${ADMIN}/smtp" "${auth[@]}" -d "{
  \"description\":   \"Resend\",
  \"senderAddress\": \"${FROM}\",
  \"senderName\":    \"${NAME}\",
  \"tls\":           true,
  \"host\":          \"smtp.resend.com:587\",
  \"user\":          \"resend\",
  \"password\":      \"${KEY}\"
}")
echo "${resp}"

ID=$(printf '%s' "${resp}" | sed -n 's/.*"id"[: ]*"\([^"]*\)".*/\1/p')
[ -n "${ID}" ] || { echo "no SMTP id in response" >&2; exit 1; }

echo "→ activating SMTP provider ${ID}"
curl -fsS -X POST "${ADMIN}/smtp/${ID}/_activate" "${auth[@]}" -d '{}' >/dev/null

echo "✓ Resend SMTP active. Send a test invite to verify delivery."
