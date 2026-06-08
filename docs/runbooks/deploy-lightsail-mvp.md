# Deploy Runbook — live-rack on AWS Lightsail

End-to-end deployment of live-rack to a shared-host AWS Lightsail instance.
Target audience: solo operator running the demo for investors.

**Zero credentials or secrets in this document. Placeholders only.**

---

## 1. Topology

```
┌──────────────────────────────────────────────────────────────────────────┐
│ AWS Lightsail · Amazon Linux 2023 · 4GB RAM · 2 vCPU · 80GB SSD          │
│ Static IP: <PUBLIC_IP>                                                   │
│                                                                          │
│  ┌── shared infra (existed before live-rack) ─────────────────────────┐  │
│  │ caddy-caddy-1            reverse proxy + TLS (network: caddy)      │  │
│  │ infra-postgres-1         used by pharmatrack, NOT by live-rack     │  │
│  │ infra-redis-1            reused by live-rack via REDIS_PREFIX      │  │
│  │ portofolio-portofolio-1  another tenant on the box                 │  │
│  │ infra-web-1 (pharmatrack)                                          │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌── live-rack (managed by deploy/docker/docker-compose.prod.yml) ────┐  │
│  │ liverack-web-1       nginx SPA            → caddy net              │  │
│  │ liverack-api-1       Go HTTP + WS         → caddy + internal       │  │
│  │ liverack-zitadel-1   OIDC IdP             → caddy + internal       │  │
│  │ liverack-zitadel-db-1                                              │  │
│  │ liverack-ingest-1    NATS → Postgres worker                        │  │
│  │ liverack-insight-1   rule engine                                   │  │
│  │ liverack-rollup-1    cron analytics                                │  │
│  │ liverack-signals-1   external feed adapters                        │  │
│  │ liverack-postgres-1  TimescaleDB (owned, separate from infra-pg)   │  │
│  │ liverack-nats-1                                                    │  │
│  │ liverack-minio-1                                                   │  │
│  └────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

Total memory budget ≈ 2.4GB. Requires 4GB swap on the host.

---

## 2. Service dependency graph

Boot order is enforced via `depends_on` health gates in `docker-compose.prod.yml`.

```mermaid
flowchart LR
  caddy[caddy-caddy-1<br/>shared]
  web[liverack-web]
  api[liverack-api]
  zit[liverack-zitadel]
  zitdb[(liverack-zitadel-db)]
  pg[(liverack-postgres)]
  nats[(liverack-nats)]
  minio[(liverack-minio)]
  redis[(infra-redis-1<br/>shared)]
  ingest[liverack-ingest]
  insight[liverack-insight]
  rollup[liverack-rollup]
  signals[liverack-signals]

  caddy --> web
  caddy --> api
  caddy --> zit

  zit --> zitdb
  api --> pg
  api --> nats
  api --> redis
  api --> minio
  api --> zit

  ingest --> pg
  ingest --> nats
  insight --> pg
  insight --> nats
  rollup --> pg
  signals --> nats

  goose_migrator[goose<br/>transient] -.runs once.-> pg
```

External dependencies:

| Dep | Used by | Failure mode |
|---|---|---|
| GHCR (ghcr.io) | docker pull | deploy step 5xx if outage |
| Resend SMTP (or equivalent) | zitadel | password resets / invites fail |
| Cloudflare DNS | TLS issuance | unreachable subdomain |
| Lightsail firewall port 443/80 | Caddy | TLS handshake fails |
| Lightsail firewall port 22 | CI deploy | SSH timeout |

---

## 3. Prerequisites

### 3.1 On the Lightsail host (one-time)

| Item | How to verify |
|---|---|
| Amazon Linux 2023 instance running | Lightsail console |
| Static IP attached | Lightsail console → Networking |
| Port 80, 443, 22 open in IPv4 Firewall | Lightsail console → Networking |
| Docker + Docker Compose v2 installed | `docker --version && docker compose version` |
| 4GB swap file active | `swapon --show` |
| Caddy already running at `~/caddy/docker-compose.yml` | `docker ps \| grep caddy` |
| `caddy` Docker network exists | `docker network ls \| grep caddy` |
| `infra-redis-1` on the `caddy` network | `docker network inspect caddy \| grep infra-redis-1` |
| Cert files at `~/caddy/certs/mrohadi.com.{pem,key}` | `ls ~/caddy/certs/` |

To add Redis to the `caddy` network if missing:
```bash
docker network connect caddy infra-redis-1
```

To create the swap file if missing:
```bash
sudo dd if=/dev/zero of=/swapfile bs=1M count=4096
sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
sudo sysctl vm.swappiness=10
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.d/99-swap.conf
```

### 3.2 On Cloudflare

| Item | Setting |
|---|---|
| DNS A record `live-rack` | Value = `<PUBLIC_IP>`, Proxy = OFF (gray cloud), TTL = Auto |

WebSocket support depends on Proxy OFF for the initial setup. Re-enable orange cloud later if desired.

### 3.3 GitHub repo

| Type | Name | Purpose |
|---|---|---|
| Secret | `SSH_HOST` | Lightsail static IP |
| Secret | `SSH_USER` | `ec2-user` |
| Secret | `SSH_PRIVATE_KEY` | Contents of Lightsail default `.pem` |
| Secret | `VITE_OIDC_ISSUER` | `https://live-rack.mrohadi.com` |
| Secret | `VITE_OIDC_CLIENT_ID` | filled after first Zitadel boot |
| Secret | `VITE_OIDC_REDIRECT_URI` | `https://live-rack.mrohadi.com/callback` |
| Secret | `VITE_API_BASE_URL` | `https://live-rack.mrohadi.com/api` |

Set via `gh secret set <name> --body "<value>"` or in the GitHub UI under Settings → Secrets and variables → Actions.

### 3.4 Local `.env.prod`

Copy `.env.prod.example` to `.env.prod` and fill **all** `CHANGE_ME_*` values.

Required keys and what they govern:

| Key | Source | Notes |
|---|---|---|
| `PUBLIC_DOMAIN` | constant | `live-rack.mrohadi.com` |
| `POSTGRES_PASSWORD` | random 32-char | drives Postgres init AND DATABASE_URL derivation |
| `ZITADEL_MASTERKEY` | random EXACTLY 32 chars | encrypts Zitadel storage; loss = bricked Zitadel |
| `ZITADEL_DB_PASSWORD` | random 32-char | Zitadel's own Postgres |
| `ZITADEL_ADMIN_EMAIL` | your email | initial admin user |
| `ZITADEL_ADMIN_PASSWORD` | strong | initial admin login |
| `OIDC_PROJECT_ID` | from Zitadel console after boot | leave blank initially, fill at step 9 |
| `S3_ACCESS_KEY` | constant | `liverack` |
| `S3_SECRET_KEY` | random | MinIO root |
| `REDIS_URL` | constant | `redis://infra-redis-1:6379` |
| `REDIS_PREFIX` | constant | `liverack:` |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USER` / `SMTP_PASSWORD` | provider | Resend, Brevo, SES, etc. |
| `SMTP_FROM` | constant | verified sender |
| `SMTP_TLS` | constant | `true` |
| `GHCR_OWNER` | constant | `mrohadi` |
| `IMAGE_TAG` | overwritten by CI | leave as `latest` |

Generate random secrets:
```bash
openssl rand -base64 32 | tr -d '/+=' | head -c 32   # 32-char alphanumeric
openssl rand -base64 24 | tr -d '/+=' | head -c 32   # 32 chars exactly for ZITADEL_MASTERKEY
```

---

## 4. First-time deploy — ordered checklist

### Step 1 — Ship `.env.prod` to host

```bash
scp .env.prod ec2-user@<PUBLIC_IP>:~/live-rack/.env.prod
ssh ec2-user@<PUBLIC_IP> 'mkdir -p ~/live-rack && chmod 600 ~/live-rack/.env.prod'
```

### Step 2 — Append Caddy snippet

```bash
ssh ec2-user@<PUBLIC_IP>
# Backup existing
cp ~/caddy/Caddyfile ~/caddy/Caddyfile.bak.$(date +%s)
# Append (or replace existing live-rack.mrohadi.com block) with content from
#   ~/live-rack/deploy/docker/caddy-snippet.live-rack.conf
# Then restart Caddy
docker compose -f ~/caddy/docker-compose.yml restart caddy
docker logs --tail 20 caddy-caddy-1   # confirm no errors
```

### Step 3 — Push code to `main`

CI runs `deploy.yml` automatically.

```bash
git push origin main
gh run watch
```

Path-filter (LR-D22) skips builds for unchanged services. First deploy builds all 6.

### Step 4 — Wait for build-push job to complete

The first run pushes 6 packages to GHCR (`liverack-web`, `liverack-api`, `liverack-ingest`, `liverack-insight`, `liverack-rollup`, `liverack-signals`). All start **private**.

### Step 5 — Flip packages to **Public**

Required because the host pulls anonymously. (Alternative: configure a `read:packages` PAT — see §6.)

1. https://github.com/mrohadi?tab=packages
2. For each of the 6 packages: click → Package settings → Danger Zone → Change visibility → Public

### Step 6 — Re-trigger deploy

```bash
gh workflow run deploy.yml --ref main
gh run watch
```

Deploy job will:
1. SCP compose file + Caddy snippet + migrations
2. Rewrite `.env.prod` to set `IMAGE_TAG=latest`, `GHCR_OWNER=<owner>`
3. `docker logout ghcr.io` to drop any stale namespace-scoped token
4. `compose pull && up -d --remove-orphans`
5. Run goose migrations via transient `golang:1.26-alpine` container against `liverack_internal` network
6. Prune dangling images
7. Smoke-check `https://<PUBLIC_DOMAIN>/api/health`

**Smoke check will fail** at this point — `OIDC_PROJECT_ID` is empty in `.env.prod`. API exits with `missing required env var: OIDC_PROJECT_ID`. Expected — continue to step 7.

### Step 7 — Bootstrap Zitadel via console

```
https://live-rack.mrohadi.com/ui/console
```

Login as `ZITADEL_ADMIN_EMAIL` with `ZITADEL_ADMIN_PASSWORD`. Forced to change password on first login.

Then:

1. Create project named `live-rack`
2. Copy the project's **Resource ID** — this becomes `OIDC_PROJECT_ID`
3. Inside the project, add a new application:
   - Name: `web`
   - Type: **User Agent / SPA**
   - Auth method: **PKCE**
   - Redirect URI: `https://live-rack.mrohadi.com/callback`
   - Post Logout URI: `https://live-rack.mrohadi.com/`
   - Development Mode: OFF
4. After creation, copy the **Client ID** — this becomes `VITE_OIDC_CLIENT_ID`
5. App → **Token Settings** tab:
   - Auth Token Type: **JWT**
   - ✓ Add user roles to access token
   - ✓ User roles inside ID token
6. Project → **Roles** tab → create 5 roles: `admin`, `manager`, `staff`, `readonly`, `service`
7. Default Settings → Login Behavior → WebAuthn → Display Name = `live-rack`

### Step 8 — Set Zitadel SMTP

```bash
ssh ec2-user@<PUBLIC_IP>
cd ~/live-rack
OIDC_ISSUER=https://live-rack.mrohadi.com \
ZITADEL_ADMIN_TOKEN=<service-user IAM_OWNER PAT> \
RESEND_API_KEY=<paste> \
SMTP_SENDER_ADDRESS=noreply@mrohadi.com \
SMTP_SENDER_NAME="live-rack" \
  bash scripts/zitadel-smtp.sh
```

To mint the service-user PAT: Console → Users → Service Users → New → generate PAT → grant role IAM_OWNER on the instance.

### Step 9 — Apply Zitadel branding

```bash
OIDC_ISSUER=https://live-rack.mrohadi.com \
ZITADEL_MGMT_TOKEN=<service-user PAT> \
  bash scripts/zitadel-branding.sh
```

For logo + Inter font: Console → Default Settings → Branding → upload manually.

### Step 10 — Update IDs in `.env.prod` and GH secrets

On host:
```bash
ssh ec2-user@<PUBLIC_IP>
cd ~/live-rack
sed -i "s|^OIDC_PROJECT_ID=.*|OIDC_PROJECT_ID=<paste-resource-id>|" .env.prod
```

Locally:
```bash
gh secret set VITE_OIDC_CLIENT_ID --body "<paste-client-id>"
```

### Step 11 — Restart impacted services

```bash
ssh ec2-user@<PUBLIC_IP>
cd ~/live-rack
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod up -d api ingest insight rollup signals
```

### Step 12 — Re-trigger deploy to rebuild `web` with the new client ID

The SPA bundle bakes `VITE_OIDC_CLIENT_ID` at **build** time. Changing the secret requires a fresh build.

```bash
gh workflow run deploy.yml --ref main
gh run watch
```

### Step 12.5 — Seed demo data (optional, one-time)

Loads the demo org `Acme Retail`, one store, 4 zones, 11 items, locations, tasks, and a Restoration pipeline. Idempotent — re-running upserts the same rows. Skip for empty-state demos.

```bash
ssh ec2-user@<PUBLIC_IP>
cd ~/live-rack
POSTGRES_PW=$(grep ^POSTGRES_PASSWORD .env.prod | cut -d= -f2-)
DB_URL="postgres://postgres:${POSTGRES_PW}@postgres:5432/liverack?sslmode=disable"

docker run --rm \
  --network liverack_liverack_internal \
  -v "$HOME/live-rack/scripts/seed:/seed" \
  -w /seed \
  -e DATABASE_URL="$DB_URL" \
  golang:1.26-alpine \
  sh -c "apk add --no-cache git >/dev/null && go mod tidy && go run ."
```

Expected output:

```
INFO seed complete org="Acme Retail" store="Store #14"
```

To **reset** seeded data before re-seeding:

```bash
docker exec liverack-postgres-1 \
  psql -U postgres -d liverack -c \
  "TRUNCATE zones, items, item_locations, tasks, pipelines, pipeline_stages, pipeline_cards, integrations, scan_events, sales_events CASCADE;"
```

> **Note:** Seed runs against the prod Postgres on the `liverack_internal` network. `scripts/seed/` is SCP'd to the host as part of every deploy if added to the workflow's `scp source:` list (currently it's not — host already has a copy from the first deploy via `git clone`, or copy manually with `scp -r scripts/seed ec2-user@HOST:~/live-rack/scripts/`).

---

### Step 12.6 — Hide Zitadel watermark + force English UI

The hosted login page ships with a `Powered by ZITADEL` footer and auto-picks the browser's Accept-Language (often Indonesian for Asia hosts). Both are policy-driven and need a service-account PAT to flip.

1. **Mint a service-account PAT** (Console UI — 60s):
   1. Open `https://live-rack.mrohadi.com/ui/console`
   2. Default Org → **Service Users** → **+ New** → Username `branding-bot` → Save
   3. On the created user: **Personal Access Tokens** → **+ New** → copy the token
   4. Default Org → **IAM** (top right org switch to "ZITADEL") → **Managers** → **+ Add** → user `branding-bot`, role `IAM_OWNER`
2. **Apply branding**:
   ```bash
   export OIDC_ISSUER=https://live-rack.mrohadi.com
   export ZITADEL_MGMT_TOKEN=<paste the PAT>
   ./scripts/zitadel-branding.sh
   ```
3. **Verify** in an incognito window — login footer shows only the live-rack copy; UI is English regardless of browser language.

> Logo and custom font upload is multipart and stays manual: Console → Default settings → Branding.

---

### Step 13 — Verify

| Check | Command |
|---|---|
| DNS | `dig +short live-rack.mrohadi.com` → `<PUBLIC_IP>` |
| TLS | `curl -I https://live-rack.mrohadi.com` → 200 |
| API health | `curl -I https://live-rack.mrohadi.com/healthz` → 200 |
| Zitadel console | open `/ui/console` in browser |
| Login page reads "live-rack" | visual |
| Passkey prompt reads "Use your passkey for live-rack" | enroll a passkey |
| Verification email arrives < 30s | trigger password reset |
| WebSocket | DevTools → Network → `wss://...` → 101 |
| Scanner camera works on mobile | open `/scanner` on phone over HTTPS |

---

## 5. Subsequent deploys (after step 13 passes once)

Just push to main:

```bash
git push origin main
```

Path filter (LR-D22) rebuilds only services whose files changed:

| You change | Rebuilds |
|---|---|
| `apps/web/**` | `liverack-web` |
| `services/api/**` | `liverack-api` |
| `services/ingest/**` | `liverack-ingest` |
| `pkg/**` or `go.work` | all 5 Go services |
| `deploy/docker/**` or `.github/workflows/deploy.yml` | all 6 |
| `migrations/**` | nothing rebuilds, but goose runs |

To force-rebuild everything: `gh workflow run deploy.yml --ref main` (the workflow_dispatch path bypasses the filter).

---

## 6. Switch packages back to private (optional, post-demo)

1. Create a fine-grained PAT with `read:packages` scope
2. `gh secret set GHCR_READ_USER --body "<your username>"`
3. `gh secret set GHCR_READ_TOKEN --body "<paste PAT>"`
4. In `.github/workflows/deploy.yml`, inside the `Pull + restart on host` step, re-add the GHCR login block (commented out today after the `docker logout ghcr.io` line):

```bash
echo "${{ secrets.GHCR_READ_TOKEN }}" | docker login ghcr.io \
  -u ${{ secrets.GHCR_READ_USER }} --password-stdin
```

5. Set each package's visibility to Private via GitHub UI.

---

## 7. Rollback

### 7.1 Roll the SPA bundle back to the previous build

Every successful build tags both `:latest` and `:<full-commit-sha>`. To pin a single service to an earlier commit:

```bash
ssh ec2-user@<PUBLIC_IP>
cd ~/live-rack
# Edit compose file inline for the one service, OR change IMAGE_TAG globally:
sed -i "s|^IMAGE_TAG=.*|IMAGE_TAG=<old-sha>|" .env.prod
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod up -d --remove-orphans
```

To restore the rolling-`latest` behavior:
```bash
sed -i "s|^IMAGE_TAG=.*|IMAGE_TAG=latest|" .env.prod
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod up -d
```

### 7.2 Roll back a DB migration

```bash
ssh ec2-user@<PUBLIC_IP>
cd ~/live-rack
DB_URL="postgres://postgres:$(grep ^POSTGRES_PASSWORD .env.prod | cut -d= -f2)@postgres:5432/liverack?sslmode=disable"
docker run --rm --network liverack_liverack_internal \
  -v "$PWD/migrations:/m:ro" \
  -e GOOSE_DRIVER=postgres -e GOOSE_DBSTRING="$DB_URL" \
  golang:1.26-alpine \
  sh -c "go install github.com/pressly/goose/v3/cmd/goose@v3.27.1 && /go/bin/goose -dir /m down"
```

### 7.3 Restore Caddy to pre-live-rack state

```bash
ssh ec2-user@<PUBLIC_IP>
ls -la ~/caddy/Caddyfile.bak.*    # pick the most recent backup
cp ~/caddy/Caddyfile.bak.<timestamp> ~/caddy/Caddyfile
docker compose -f ~/caddy/docker-compose.yml restart caddy
```

---

## 8. Troubleshooting (from real incidents)

### 8.1 `denied: denied` pulling a GHCR image

| Symptom | Root cause | Fix |
|---|---|---|
| `Head ".../mrohadi/liverack-*/manifests/<sha>": denied` | Package still private | Flip to Public (§5) |
| `Head ".../zitadel/zitadel/manifests/v2.71.9": denied` | Stale namespace-scoped token on host (`docker login ghcr.io` left a cookie) | `docker logout ghcr.io` (deploy.yml does this automatically) |
| `Head ".../mrohadi/liverack-*/manifests/fc1f2f...": denied` | Tag mismatch — old workflow used `sha-` prefix | Fixed in LR-D17 (prefix= empty) |

### 8.2 SCP `dial tcp ***:22: i/o timeout`

Lightsail firewall blocks port 22 from runner. Fix in Lightsail → Networking → IPv4 Firewall → SSH source = **Any IPv4** (or restrict to a list that includes GH Actions ranges).

### 8.3 Build fails: `ERR_UNKNOWN_BUILTIN_MODULE`

pnpm 11.x via corepack on Node 20 hits builtin module mismatch. Fixed in LR-D15 — Dockerfile uses `node:22-alpine` + pinned `pnpm@11.5.0`.

### 8.4 Build fails: `"/package.json": not found`

Repo has no root `package.json` (workspace defined in `pnpm-workspace.yaml` only). Fixed in LR-D14.

### 8.5 Build fails: GitHub Actions Cache 503

`cache-to: type=gha` is now annotated with `ignore-error=true` (LR-D19). Cache outage no longer fails the build.

### 8.6 API crashes: `missing required env var: OIDC_PROJECT_ID`

You haven't completed step 7. Boot Zitadel, copy project Resource ID into `.env.prod` on host.

### 8.7 Console: `[unknown] HTTP 405`

Caddy doesn't forward gRPC-Web paths. Fixed in LR-D23 — `/zitadel.*` added to `@zitadel` path matcher, upstream switched to `h2c://`.

### 8.8 Goose: `password authentication failed for user "postgres"`

`POSTGRES_PASSWORD` and the password in `DATABASE_URL` drifted. Fixed in LR-D21 — `DATABASE_URL` now derived inline from `POSTGRES_PASSWORD` in compose.

If volume was initialized with a stale password:
```bash
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod down postgres
docker volume rm liverack_postgres_data
docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod up -d postgres
```

### 8.9 setup-buildx times out reaching docker.io

GitHub-side flap. Fixed in LR-D16 — buildkit pinned to `moby/buildkit:v0.15.2` to improve image cache hits. If it still happens, rerun the failed matrix job: `gh run rerun <run-id> --failed`.

### 8.10 Out-of-memory kills containers

Total cap ≈ 2.4GB. With OS + other tenants (`pharmatrack`, `portofolio`) the 4GB plan is tight. If kernel OOM kills random containers, upgrade Lightsail to 8GB plan (one-click, no provision time).

---

## 9. Cost

| Item | Monthly |
|---|---|
| Lightsail instance (4GB, shared) | $0 marginal (already paid by other apps) |
| Optional 8GB upgrade | +$20 |
| Static IP (attached) | $0 |
| Object Storage backups | $1 |
| Domain | $0 (owned) |
| Resend SMTP free tier | $0 |
| GHCR storage | $0 (within free tier) |
| GitHub Actions minutes | $0 (under free 2000/mo) |
| **Marginal total** | **~$1–21/mo** |

---

## 10. Backup

Cron on host, nightly 03:00 UTC:

```bash
ssh ec2-user@<PUBLIC_IP>
crontab -e
# Add:
0 3 * * * docker exec liverack-postgres-1 pg_dump -U postgres liverack | gzip > /home/ec2-user/backups/liverack-$(date +\%F).sql.gz
0 4 * * * find /home/ec2-user/backups -name '*.sql.gz' -mtime +7 -delete
```

Push offsite via AWS CLI to Lightsail Object Storage (or S3) later.

---

## 11. Skipped from MVP scope

Documented for re-enable after demo:

- P6 POS integrations (Shopify, Square) — `services/integrations/` skeleton not built
- P7 Analytics (ClickHouse, heatmaps) — services exist but compose excludes ClickHouse
- P8 External signals (weather, transit)
- P9 RBAC + audit log surface
- ELK stack (Elasticsearch + Kibana + Filebeat) — heavy for 4GB host
- ZAP baseline scan in CI (runs weekly, not per-PR)
- Goose binary baked into a dedicated migrator image (uses transient golang container today)

---

## 12. Quick command reference

```bash
# Watch a deploy
gh run watch $(gh run list --workflow=deploy.yml --limit 1 --json databaseId --jq '.[0].databaseId')

# Re-trigger deploy manually
gh workflow run deploy.yml --ref main

# Tail logs from one service on host
ssh ec2-user@<PUBLIC_IP> 'docker logs -f --tail 100 liverack-api-1'

# Inspect compose state
ssh ec2-user@<PUBLIC_IP> 'cd ~/live-rack && docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod ps'

# Restart one service
ssh ec2-user@<PUBLIC_IP> 'cd ~/live-rack && docker compose -f deploy/docker/docker-compose.prod.yml --env-file .env.prod restart api'

# Force-rebuild every image
gh workflow run deploy.yml --ref main
```
