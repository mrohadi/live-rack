# ZAP baseline — findings & remediation

The weekly `zap-baseline` workflow runs a passive DAST scan and gates on
`.zap/rules.tsv`. This page tracks expected alerts and how we handle them.

## Gating policy

| Action | Meaning |
|---|---|
| `FAIL` | Real vulnerability class — breaks the build (XSS, SQLi, error disclosure) |
| `WARN` | Hardening item — tracked, not blocking |
| `IGNORE` | Expected/false-positive for a token-auth JSON API + SPA |

## Baseline hardening checklist

- [x] **SQL injection** — sqlc parameterised queries only; no string interpolation. (`40018` FAIL)
- [x] **Reflected XSS** — React escaping; no `dangerouslySetInnerHTML`. (`40012` FAIL)
- [x] **Error disclosure** — handlers return generic messages; stack traces never sent. (`90022` FAIL)
- [ ] **Security headers** — add `Content-Security-Policy`, `Permissions-Policy`, `X-Content-Type-Options`, HSTS at the gateway/CDN before GA. (`10063` WARN)
- [x] **AuthN** — all `/api/v1/*` require Bearer (OIDC JWT or `lrk_` service token); webhooks verify per-vendor signatures.
- [x] **Rate limiting** — gateway: 100 req/s/org, 20 scan events/s/scanner.

## Triaging a new alert

1. Reproduce locally: `docker compose up` + run the [baseline action](https://github.com/zaproxy/action-baseline) against `http://localhost:8080`.
2. Real issue → fix + add a regression test; keep the rule at `FAIL`.
3. False positive → set the plugin id to `IGNORE` in `.zap/rules.tsv` with a one-line justification.

## Pen test

Pre-GA, commission an external pen test covering multi-tenant isolation (RLS
bypass attempts across `org_id`/zone scope), service-token handling, and webhook
signature spoofing. File findings as `SEV`-tagged issues and link fixes here.
