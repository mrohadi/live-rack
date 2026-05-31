# Runbook: analytics ingest stall

**Symptom:** Analytics screen stale; ClickHouse row counts not advancing; NATS
consumer `ingest` shows growing pending.

## Triage
1. NATS: `curl -s localhost:8222/jsz?consumers=1` — inspect `LIVE_RACK` consumer `num_pending`.
2. Ingest worker logs: look for `sink pos.sale` / `sink scan.recorded` errors.
3. ClickHouse reachable: `curl -s -u "$CH_AUTH" 'localhost:8123/?query=SELECT%201'` (creds from `CLICKHOUSE_URL`).
4. Confirm MV targets advancing: `SELECT max(bucket) FROM zone_perf_5m`.

## Mitigate
- **ClickHouse down/unreachable:** restart `clickhouse` container; the worker runs `Migrate` on boot and resumes from NATS (24h retention).
- **Worker crashed:** restart `services/ingest`; durable JetStream replays unacked messages.
- **Schema drift:** re-run `chstore.Migrate` (idempotent `CREATE ... IF NOT EXISTS`).
- **Rollup jobs failing:** check `services/rollup` logs; jobs are idempotent (ReplacingMergeTree) — safe to re-run for a day.

## Verify
- Consumer `num_pending` trending to 0.
- `SELECT count() FROM scan_events_raw` increasing; Analytics screen current.

## Escalate
- Sustained backlog > NATS retention (24h) risks data loss → SEV2, page primary.
