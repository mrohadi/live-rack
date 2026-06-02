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
    typeof f.suggested_task === "string" && typeof f.title === "string" && typeof f.id === "string"
  );
}

/** Why a task notification fired. Mirrors events.TaskNotifyKind. */
export type TaskNotifyKind = "assigned" | "deadline";

/** Task notification pushed over the WS stream (events.TaskNotified). */
export interface TaskNotification {
  org_id: string;
  store_id: string;
  task_id: string;
  assignee_id: string;
  title: string;
  due_at?: string | null;
  kind: TaskNotifyKind;
  ts: string;
}

/** Type guard: does a decoded WS frame look like a TaskNotification? Pure. */
export function isTaskNotification(frame: unknown): frame is TaskNotification {
  if (typeof frame !== "object" || frame === null) return false;
  const f = frame as Record<string, unknown>;
  return (
    typeof f.task_id === "string" &&
    typeof f.title === "string" &&
    (f.kind === "assigned" || f.kind === "deadline")
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

// openTaskNotificationSocket maintains one auto-reconnecting socket and invokes
// onNotify for frames that are task notifications. Returns a disposer.
export function openTaskNotificationSocket(
  getToken: () => Promise<string | null>,
  onNotify: (n: TaskNotification) => void,
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
        if (isTaskNotification(frame)) onNotify(frame);
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
