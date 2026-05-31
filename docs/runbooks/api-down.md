# Runbook: API down / 5xx spike

**Symptom:** `/healthz` returns non-200, or 5xx error rate > 1% on the Ops dashboard.

## Triage
1. Check health: `curl -s $API/healthz` (expect `{"status":"ok"}`; `{"db":"down"}` = DB issue → see [postgres-failover.md](postgres-failover.md)).
2. Check pods/containers: `docker compose -f deploy/docker/docker-compose.yml ps` (or orchestrator).
3. APM: Kibana → *Ops Overview* → error rate panel; drill into traces for the failing route.
4. Recent deploy? Check the last release tag / commit.

## Mitigate
- **Bad deploy:** roll back to the previous image tag; redeploy.
- **DB saturation:** check `pg_stat_activity`; kill long queries; scale connections.
- **Dependency down (NATS/ClickHouse/Redis):** restart the dependency; API degrades gracefully for reads.
- **Overload:** confirm gateway rate limits (100 req/s/org, 20 scans/s/scanner) are active.

## Verify
- `/healthz` green for 5 min.
- 5xx rate back < 1%; p95 latency < 250ms.

## Escalate
- DB data loss suspected → SEV1, page secondary + DBA.
- Unresolved after 30 min → escalate per [on-call.md](on-call.md).
