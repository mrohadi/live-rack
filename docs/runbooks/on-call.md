# On-call rotation

## Rotation
- **Weekly**, Mon 10:00 → Mon 10:00 (local). Primary + secondary each week.
- Handover at week start: review open incidents, deploys in flight, silenced alerts.
- Secondary backs up the primary (no ack in 10 min → secondary paged).

## Responsibilities
- Acknowledge SEV1/SEV2 pages within **5 minutes**.
- Drive incident to mitigation; open an incident channel for SEV1/SEV2.
- File a postmortem within **48h** for any SEV1 or customer-impacting SEV2.

## Comms
- Incident channel: `#inc-<date>-<short-desc>`.
- Status updates every 30 min during an active SEV1.
- Customer-facing status page updated for tenant-impacting incidents.

## Paging matrix
| Severity | Who | When |
|---|---|---|
| SEV1 | Primary + Secondary | Immediately |
| SEV2 | Primary | Immediately |
| SEV3 | Primary | Next business day |

## Useful entrypoints
- Health: `GET /healthz` · Metrics: `GET /metrics`
- Dashboards: Kibana → *live-rack · Ops Overview*
- Runbooks: [README.md](README.md)
