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

const WS_BASE = (import.meta.env.VITE_API_URL ?? "http://localhost:8080").replace(/^http/, "ws");

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
