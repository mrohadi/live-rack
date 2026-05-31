# Runbook: webhook signature rejections

**Symptom:** Integrations → webhook log shows `rejected`; partners report missing
sales; spike in 401/400 on `/webhooks/*`.

## Triage
1. Identify vendor: Shopify / Square / Stripe / Shippo.
2. Inspect rejection reason in webhook log + API logs.
3. Confirm the per-vendor secret matches the partner dashboard.

## Common causes & fixes
- **Stripe:** `Stripe-Signature` HMAC over `"<t>.<body>"`. Rejections usually mean a stale `whsec`. Rotate and update the org's secret. Clock skew on `t` also matters.
- **Shopify:** base64 HMAC of raw body. Body must be the **raw** bytes (no re-encoding by a proxy).
- **Shippo:** shared `X-Shippo-Token` mismatch — re-sync the configured token.
- **Body mutation:** any proxy that reformats JSON breaks HMAC. Ensure raw passthrough.

## Mitigate
- Update the affected secret; ask the partner to redeliver recent events (Shopify/Stripe support replay).
- Idempotency via `EventID` means safe replays (no duplicate sales).

## Verify
- New deliveries show `processed`; sales appear in the dashboard.

## Escalate
- Vendor-wide outage → SEV3, monitor partner status page.
