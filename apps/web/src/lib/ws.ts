export interface ScanRecorded {
  org_id: string;
  store_id: string;
  zone_id: string;
  scanner_id: string;
  sku: string;
  action: string;
  valid: boolean;
  reason?: string;
  ts: string;
}

type Handler = (ev: ScanRecorded) => void;

/** Recommendation pushed by the insight engine over the signal stream. */
export interface Recommendation {
  id: string;
  org_id: string;
  store_id: string;
  kind: string;
  severity: string;
  title: string;
  rationale: string;
  suggested_task: string;
  created_at: string;
}

/** Type guard: does a decoded WS frame look like a Recommendation? Pure. */
export function isRecommendation(frame: unknown): frame is Recommendation {
  if (typeof frame !== "object" || frame === null) return false;
  const f = frame as Record<string, unknown>;
  return (
    typeof f.suggested_task === "string" &&
    typeof f.title === "string" &&
    typeof f.id === "string"
  );
}

const WS_BASE = (import.meta.env.VITE_API_URL ?? "http://localhost:8080").replace(/^http/, "ws");

// openRecommendationSocket maintains one auto-reconnecting socket and invokes
// onRec for frames that are recommendations. Returns a disposer.
export function openRecommendationSocket(
  getToken: () => Promise<string | null>,
  onRec: (rec: Recommendation) => void,
): () => void {
  let socket: WebSocket | null = null;
  let closed = false;
  let retry = 0;

  const scheduleReconnect = () => {
    if (closed) return;
    const delay = Math.min(1000 * 2 ** retry, 30_000);
    retry += 1;
    setTimeout(() => void connect(), delay);
  };

  const connect = async () => {
    if (closed) return;
    const token = await getToken();
    if (!token) return scheduleReconnect();

    socket = new WebSocket(`${WS_BASE}/api/v1/ws?token=${encodeURIComponent(token)}`);
    socket.onopen = () => {
      retry = 0;
    };
    socket.onmessage = (e) => {
      try {
        const frame: unknown = JSON.parse(e.data as string);
        if (isRecommendation(frame)) onRec(frame);
      } catch {
        // ignore malformed frames
      }
    };
    socket.onclose = () => scheduleReconnect();
    socket.onerror = () => socket?.close();
  };

  void connect();
  return () => {
    closed = true;
    socket?.close();
  };
}

// openScanSocket maintains one auto-reconnecting socket; returns a disposer.
export function openScanSocket(
  getToken: () => Promise<string | null>,
  onEvent: Handler,
): () => void {
  let socket: WebSocket | null = null;
  let closed = false;
  let retry = 0;

  const scheduleReconnect = () => {
    if (closed) return;
    const delay = Math.min(1000 * 2 ** retry, 30_000);
    retry += 1;
    setTimeout(() => void connect(), delay);
  };

  const connect = async () => {
    if (closed) return;
    const token = await getToken();
    if (!token) return scheduleReconnect();

    socket = new WebSocket(`${WS_BASE}/api/v1/ws?token=${encodeURIComponent(token)}`);
    socket.onopen = () => {
      retry = 0;
    };
    socket.onmessage = (e) => {
      try {
        onEvent(JSON.parse(e.data as string) as ScanRecorded);
      } catch {
        // ignore malformed frames
      }
    };
    socket.onclose = () => scheduleReconnect();
    socket.onerror = () => socket?.close();
  };

  void connect();
  return () => {
    closed = true;
    socket?.close();
  };
}
