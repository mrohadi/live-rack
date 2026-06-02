// k6 scenario: sustain 10k scan ingests/min (~167 rps) for 5 minutes.
//
//   k6 run -e LOAD_TOKEN=... -e LOAD_STORE_ID=... -e LOAD_ZONE_ID=... load/scans.js
import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";
import { BASE_URL, STORE_ID, authHeaders, scanPayload } from "./lib/config.js";

const scansSent = new Counter("scans_sent");

export const options = {
  scenarios: {
    scans: {
      executor: "constant-arrival-rate",
      rate: 167, // ~10,020 per minute
      timeUnit: "1s",
      duration: "5m",
      preAllocatedVUs: 100,
      maxVUs: 400,
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"], // <1% errors
    http_req_duration: ["p(95)<250"], // 95th pct under 250ms
  },
};

export default function () {
  const url = `${BASE_URL}/api/v1/stores/${STORE_ID}/scans`;
  const res = http.post(url, JSON.stringify(scanPayload(__ITER)), { headers: authHeaders() });
  check(res, { "scan accepted": (r) => r.status === 200 || r.status === 201 });
  scansSent.add(1);
}
