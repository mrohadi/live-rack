# live-rack runbooks

Operational playbooks for on-call. Each runbook follows: **Symptom → Triage →
Mitigate → Verify → Escalate**.

| Runbook | Trigger |
|---|---|
| [api-down.md](api-down.md) | `/healthz` failing, 5xx spike |
| [ingest-stall.md](ingest-stall.md) | ClickHouse analytics lagging / NATS backlog |
| [postgres-failover.md](postgres-failover.md) | Primary DB unhealthy |
| [webhook-failures.md](webhook-failures.md) | POS/Stripe webhook signature rejections |
| [on-call.md](on-call.md) | Rotation, severities, comms |
| [zitadel-setup.md](zitadel-setup.md) | Auth/OIDC deploy + `ZITADEL_MGMT_TOKEN` |

## Severities

| Sev | Definition | Response |
|---|---|---|
| SEV1 | Full outage or data loss | Page primary + secondary immediately |
| SEV2 | Major feature degraded, tenant-impacting | Page primary |
| SEV3 | Minor/partial, no customer impact | Next business day |

## Golden signals (Kibana → *Ops Overview*)

- Scan ingest rate (target ~10k/min sustainable)
- API `http_req_duration` p95 < 250ms
- 5xx error rate < 1%
- NATS consumer pending / ClickHouse insert lag
