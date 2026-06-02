// Shared config for the k6 load suite. Pure helpers (no k6 runtime imports) so
// they can be linted/checked with `node --check`.

export const BASE_URL = __ENV.LOAD_BASE_URL || "http://localhost:8080";
export const WS_URL = (__ENV.LOAD_WS_URL || BASE_URL).replace(/^http/, "ws");
export const TOKEN = __ENV.LOAD_TOKEN || "";
export const ORG_ID = __ENV.LOAD_ORG_ID || "00000000-0000-0000-0000-000000000000";
export const STORE_ID = __ENV.LOAD_STORE_ID || "00000000-0000-0000-0000-000000000000";
export const ZONE_ID = __ENV.LOAD_ZONE_ID || "00000000-0000-0000-0000-000000000000";

// authHeaders builds the Bearer auth + JSON headers. Pure.
export function authHeaders() {
  return { Authorization: `Bearer ${TOKEN}`, "Content-Type": "application/json" };
}

// scanPayload builds a deterministic scan body for the given iteration. Pure.
export function scanPayload(i) {
  return {
    store_id: STORE_ID,
    zone_id: ZONE_ID,
    scanner_id: `k6-${i % 50}`,
    sku: `LR-${1000 + (i % 200)}`,
    action: "pick",
  };
}
