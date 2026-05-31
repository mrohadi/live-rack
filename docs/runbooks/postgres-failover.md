# Runbook: Postgres failover

**Symptom:** `/healthz` returns `{"db":"down"}`; writes failing; `pg_isready` fails.

## Triage
1. `pg_isready -h $PGHOST` and check primary container/instance health.
2. Disk/connection exhaustion: `SELECT count(*) FROM pg_stat_activity;` vs `max_connections`.
3. Replication lag (HA): `SELECT * FROM pg_stat_replication;`.

## Mitigate
- **Connection exhaustion:** terminate idle/long txns; lower app pool max; restart API to reset pools.
- **Primary down (HA):** promote a healthy replica; repoint `DATABASE_URL`; restart API + workers.
- **Disk full:** free WAL/space; never drop tenant data; expand volume.
- After failover, confirm RLS session vars still set (`app.org_id`, `app.user_id`) — they are per-connection.

## Verify
- `/healthz` green; a tenant read/write round-trips.
- Replication re-established; lag < 10s.

## Escalate
- Suspected data loss or corruption → SEV1, page DBA + secondary.
