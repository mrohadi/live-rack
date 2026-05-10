---
description: Seed local dev database from references/live-rack/project/data.jsx fixture data
allowed-tools: Read, Bash(make:*), Bash(go:*), Bash(psql:*)
---

# Seed Dev Database

## Steps

1. Verify dev stack running: `make dev-status`
2. Read fixture source: `references/live-rack/project/data.jsx`
3. Run seed command: `make seed`

## What Gets Seeded

From `window.LR_DATA`:
- **Zones**: 11 zones (A1–C3) into `zones` table for org `dev-org`
- **Items**: 10 SKUs into `items` + `item_locations`
- **Tasks**: 6 tasks into `tasks`
- **Pipeline**: Item Restoration pipeline with 5 stages, 8 cards
- **Integrations**: 9 integrations (connected/paused/available status)
- **Scan events**: 50 synthetic scan events spread across past 24h (TimescaleDB)
- **Sales events**: 24h sparkline data into `sales_events` hypertable

## Verify

```bash
psql $DATABASE_URL -c "SELECT id, name, type FROM zones WHERE org_id = 'dev-org-id' ORDER BY id;"
psql $DATABASE_URL -c "SELECT COUNT(*) FROM scan_events WHERE ts > now() - interval '1 day';"
```

## Reset

```bash
make seed-reset  # Truncates all seeded data, re-runs seed
```
