# Load suite (k6)

Targets the MVP SLOs: **10k scans/min** ingest and **1k concurrent WS** clients.

## Prereqs

- [k6](https://k6.io/docs/get-started/installation/) installed.
- A running API (`make dev` + `go run ./services/api`) and a valid access token.

## Env

| Var | Default | Notes |
|---|---|---|
| `LOAD_BASE_URL` | `http://localhost:8080` | API base |
| `LOAD_WS_URL` | `LOAD_BASE_URL` | WS base (http→ws auto) |
| `LOAD_TOKEN` | — | Bearer JWT or `lrk_` service token |
| `LOAD_STORE_ID` / `LOAD_ZONE_ID` | zero UUID | scan target scope |

## Run

```bash
# 10k scans/min for 5m
k6 run -e LOAD_TOKEN=$TOKEN -e LOAD_STORE_ID=$STORE -e LOAD_ZONE_ID=$ZONE load/scans.js

# 1k concurrent WebSocket clients for 3m
k6 run -e LOAD_TOKEN=$TOKEN load/ws.js
```

## Thresholds (CI-gating)

- `scans.js`: `http_req_failed < 1%`, `http_req_duration p95 < 250ms`.
- `ws.js`: `ws_connecting p95 < 1s`, `ws_session_duration p90 > 30s`.

A non-zero exit means a threshold breached.
