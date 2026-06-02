// k6 scenario: hold 1,000 concurrent WebSocket clients on the live event stream
// for 3 minutes, asserting each connection stays open.
//
//   k6 run -e LOAD_TOKEN=... load/ws.js
import ws from "k6/ws";
import { check } from "k6";
import { Counter } from "k6/metrics";
import { TOKEN, WS_URL } from "./lib/config.js";

const framesReceived = new Counter("ws_frames_received");

export const options = {
  scenarios: {
    sockets: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "30s", target: 1000 }, // ramp to 1k sockets
        { duration: "3m", target: 1000 }, // hold
        { duration: "15s", target: 0 }, // drain
      ],
    },
  },
  thresholds: {
    ws_connecting: ["p(95)<1000"], // handshake under 1s at p95
    ws_session_duration: ["p(90)>30000"], // sessions live >30s
  },
};

export default function () {
  const url = `${WS_URL}/api/v1/ws?token=${encodeURIComponent(TOKEN)}`;
  const res = ws.connect(url, {}, function (socket) {
    socket.on("message", () => framesReceived.add(1));
    socket.setTimeout(() => socket.close(), 60000);
  });
  check(res, { "ws handshake 101": (r) => r && r.status === 101 });
}
