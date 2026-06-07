# MVP Deploy Strategy — AWS Lightsail (shared host)

Target: investor demo. Constraints: low budget, single region, single operator.
Scope skips P6 (POS integrations) and P7–P11.

This doc reflects the **shared-host** topology — Lightsail already runs Caddy
+ Postgres + Redis for two other projects (pharmatrack, portofolio). live-rack
joins the same host and reuses Caddy + Redis.

---

## 1. Topology

```
┌──────────────────────────────────────────────────────────────────┐
│ Lightsail instance — Amazon Linux 2023, 4GB / 2vCPU / 80GB SSD   │
│ Static IP 18.140.191.209                                         │
│                                                                  │
│  ┌────────────────────────── existing infra ─────────────────┐   │
│  │ caddy-caddy-1 (network: caddy) ── shared reverse proxy    │   │
│  │ infra-postgres-1 ── shared, NOT used by live-rack         │   │
│  │ infra-redis-1    ── reused, namespaced via REDIS_PREFIX   │   │
│  │ portofolio-portofolio-1, infra-web-1 (pharmatrack)        │   │
│  └───────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────── live-rack (this compose) ─────────────┐   │
│  │ web (nginx SPA)         ── joins `caddy` network          │   │
│  │ api (Go)                ── `caddy` + `liverack_internal`  │   │
│  │ zitadel (OIDC)          ── `caddy` + `liverack_internal`  │   │
│  │ ingest / insight / rollup / signals (Go workers)          │   │
│  │ postgres (TimescaleDB, owned)                             │   │
│  │ zitadel-db (postgres:16-alpine)                           │   │
│  │ nats (JetStream)                                          │   │
│  │ minio                                                     │   │
│  └───────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
         │
         ▼ Object Storage (5GB, $1/mo) — nightly pg_dump
```

Memory cap total ≈ 2.4GB. With OS + other projects, fits in 4GB **only with 4GB swap**.

Skipped from MVP: ClickHouse, ELK, APM. Reintroduce post-funding.

---

## 2. Pre-deploy host prep (one-time, SSH)

```bash
ssh ec2-user@18.140.191.209

# 1. Swap
sudo dd if=/dev/zero of=/swapfile bs=1M count=4096
sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
sudo sysctl vm.swappiness=10
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.d/99-swap.conf

# 2. Repo location
mkdir -p ~/live-rack
cd ~/live-rack

# 3. Confirm Caddy network exists (it already does on this host)
docker network ls | grep -w caddy

# 4. Add live-rack subdomain to existing Caddyfile
cat ~/live-rack/deploy/docker/caddy-snippet.live-rack.conf | sudo tee -a ~/caddy/Caddyfile
docker compose -f ~/caddy/docker-compose.yml restart caddy

# 5. Cloudflare DNS — add A record live-rack.mrohadi.com -> 18.140.191.209, proxy OFF
```

---

## 3. Secrets (.env.prod)

Copy `.env.prod.example` from the repo, fill secrets, ship via scp:

```bash
# Local
cp .env.prod.example .env.prod
$EDITOR .env.prod   # fill every CHANGE_ME

scp .env.prod ec2-user@18.140.191.209:~/live-rack/.env.prod
ssh ec2-user@18.140.191.209 'chmod 600 ~/live-rack/.env.prod'
```

Required from CI (set as GitHub Action secrets):
- `SSH_HOST`, `SSH_USER`, `SSH_PRIVATE_KEY`
- `GHCR_READ_USER`, `GHCR_READ_TOKEN` (PAT with `read:packages`)

Required as GitHub **vars** (Settings → Variables → Actions):
- `VITE_OIDC_ISSUER`, `VITE_OIDC_CLIENT_ID`, `VITE_OIDC_REDIRECT_URI`, `VITE_API_BASE_URL`

---

## 4. CI/CD pipeline

`.github/workflows/deploy.yml` — triggers on push to `main` or manual.

Stages:
1. **build-push** — matrix builds 6 images (web + 5 Go services), pushes to GHCR
   with sha + `latest` tags. Uses gha cache for layer reuse.
2. **deploy** — scp compose + caddy snippet + migrations to host, SSH:
   - inject `IMAGE_TAG=<sha>` into `.env.prod`
   - `docker login ghcr.io`
   - `docker compose pull && up -d --remove-orphans`
   - run goose migrations against live Postgres
   - prune dangling images
3. **smoke** — curl `https://live-rack.mrohadi.com/api/health` with retries.

Concurrency group `deploy-prod` ensures only one deploy at a time, never cancels.

---

## 5. Zitadel branding + SMTP

### Branding (live-rack name on login + passkey screens)

Set via env in `docker-compose.prod.yml` (already wired):
- `ZITADEL_DEFAULTINSTANCE_INSTANCENAME=live-rack`
- `ZITADEL_FIRSTINSTANCE_INSTANCENAME=live-rack`
- `ZITADEL_WEBAUTHN_DISPLAYNAME=live-rack` ← passkey browser prompt reads
  "Use your passkey for **live-rack**"

After first boot, mint an IAM_OWNER PAT and apply visual branding:
```bash
ssh ec2-user@host
cd ~/live-rack
OIDC_ISSUER=https://live-rack.mrohadi.com \
ZITADEL_MGMT_TOKEN=<service-user PAT> \
  bash scripts/zitadel-branding.sh
```

Logo + Inter font upload manual: Console → Default settings → Branding.

### SMTP (verification + reset emails)

Recommend **Resend** (free 3k/mo). Get key, then:

```bash
OIDC_ISSUER=https://live-rack.mrohadi.com \
ZITADEL_ADMIN_TOKEN=<IAM_OWNER PAT> \
RESEND_API_KEY=re_xxx \
SMTP_SENDER_ADDRESS=noreply@mrohadi.com \
SMTP_SENDER_NAME="live-rack" \
  bash scripts/zitadel-smtp.sh
```

Verify by triggering a password reset; check Resend dashboard for delivery.

---

## 6. Deploy steps (first cutover)

```bash
# Local — push code
git push origin main          # triggers .github/workflows/deploy.yml

# Watch
gh run watch
```

Manual fallback (if Actions fails):

```bash
ssh ec2-user@18.140.191.209
cd ~/live-rack
echo "$GHCR_READ_TOKEN" | docker login ghcr.io -u $GHCR_READ_USER --password-stdin
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod pull
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod up -d
```

Post-cutover (one-time):
```bash
# Zitadel Console — https://live-rack.mrohadi.com/ui/console
# 1. Login admin (creds from .env.prod)
# 2. Create project "live-rack" → copy Project ID
# 3. Add SPA app "web" (PKCE) → redirect https://live-rack.mrohadi.com/callback
#    → copy Client ID
# 4. App Token Settings: JWT + user roles in access token + ID token
# 5. Roles: admin, manager, staff, readonly, service
# 6. Default Settings → Login Behavior → WebAuthn → Display Name = "live-rack"
#
# Update GitHub vars:
#   VITE_OIDC_CLIENT_ID = <client id>
#   VITE_API_BASE_URL = https://live-rack.mrohadi.com/api
#   VITE_OIDC_ISSUER = https://live-rack.mrohadi.com
#   VITE_OIDC_REDIRECT_URI = https://live-rack.mrohadi.com/callback
#
# Update .env.prod on host with OIDC_PROJECT_ID, then re-trigger deploy.
```

---

## 7. Backup

Cron on host, nightly 03:00 UTC:

```bash
crontab -e
# Add:
0 3 * * * docker exec liverack-postgres-1 pg_dump -U postgres liverack | gzip | aws s3 cp - s3://liverack-backup/$(date +\%F).sql.gz
```

S3 lifecycle: delete after 7 days. ~5GB free tier easily covers.

---

## 8. Smoke checklist

- [ ] `dig +short live-rack.mrohadi.com` → 18.140.191.209
- [ ] `curl -I https://live-rack.mrohadi.com` → 200
- [ ] `curl -I https://live-rack.mrohadi.com/api/health` → 200
- [ ] `curl -I https://live-rack.mrohadi.com/ui/console` → 200 (Zitadel)
- [ ] Sign-in page reads "live-rack"
- [ ] Add passkey: browser prompt reads "Use your passkey for live-rack"
- [ ] Verification email lands within 30s
- [ ] WebSocket connects (DevTools Network → ws:// → 101)
- [ ] Camera works in scanner (HTTPS required)
- [ ] Lightsail snapshot taken 1hr before demo

---

## 9. Cost

| Item | Monthly |
|---|---|
| Lightsail instance (shared) | $0 marginal |
| Static IP | $0 |
| Object Storage backups | $1 |
| Domain | $0 (owned) |
| Resend SMTP | $0 (free tier) |
| GHCR storage | $0 (public/small private) |
| GitHub Actions minutes | $0 (under free 2000) |
| **Total** | **~$1/mo** |

If instance bumped to 8GB for headroom: +$20/mo = $21/mo total.

---

## 10. CI pipeline audit (current state)

| Workflow | Trigger | Status |
|---|---|---|
| `ci.yml` | push feat/phase/fix/chore, PR | ✅ passes |
| `security.yml` | push main/phase, PR, weekly | ✅ fixed (LR-D07) |
| `zap-baseline.yml` | weekly, manual | ✅ lightened |
| `deploy.yml` | push main, manual | ✅ new (LR-D06) |

Pipeline fixes shipped this phase:
- **security.yml** Trivy step had wrong Dockerfile path → fixed to
  `services/api/Dockerfile`
- **security.yml** gosec matrix only covered 3 modules → expanded to all
  9 Go modules
- **zap-baseline.yml** brought up ClickHouse unnecessarily → dropped

---

## 11. Risks + mitigations

| Risk | Mitigation |
|---|---|
| OOM (4GB tight) | 4GB swap + mem_limits cap stack at 2.4GB |
| Caddy restart breaks other apps | Snippet appended idempotently; restart isolates to Caddy container only |
| DB migration mid-deploy | goose runs after compose up; rollback via `goose down` if needed |
| GHCR pull rate-limited | Use authenticated pull on host (token in deploy workflow) |
| Cloudflare proxy breaks WS | DNS-only mode (gray cloud) for `live-rack` |
| Investor wifi blocks 443 | Hotspot fallback tested |
| Zitadel masterkey loss | Backup `.env.prod` to password manager |

---

## 12. Rollback

```bash
ssh ec2-user@18.140.191.209
cd ~/live-rack
# Identify previous good sha from `docker images | grep liverack`
sed -i "s/^IMAGE_TAG=.*/IMAGE_TAG=<previous-sha>/" .env.prod
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod up -d
```

Or trigger `deploy.yml` via `workflow_dispatch` on a previous commit.

---

## 13. Post-funding restore list

Reintroduce in this order:
1. ClickHouse + ingest worker (P7)
2. Rollup jobs (P7)
3. Move Postgres to managed RDS / Neon
4. ELK or Grafana Cloud (free tier)
5. P6 POS integrations (Shopify, Square)
6. Multi-region failover
