import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { openScanSocket, type ScanRecorded } from "../ws";

class FakeWebSocket {
  static last: FakeWebSocket | null = null;
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  close = vi.fn();
  constructor(public url: string) {
    FakeWebSocket.last = this;
  }
}

beforeEach(() => {
  vi.stubGlobal("WebSocket", FakeWebSocket as unknown as typeof WebSocket);
});
afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("openScanSocket", () => {
  it("parses messages and forwards typed events", async () => {
    const events: ScanRecorded[] = [];
    const close = openScanSocket(
      async () => "tok",
      (e) => events.push(e),
    );

    await vi.waitFor(() => expect(FakeWebSocket.last).not.toBeNull());
    expect(FakeWebSocket.last!.url).toContain("token=tok");

    FakeWebSocket.last!.onmessage?.({
      data: JSON.stringify({ zone_id: "z1", sku: "X", valid: true, action: "place", ts: "t" }),
    });
    expect(events).toHaveLength(1);
    expect(events[0].sku).toBe("X");

    close();
    expect(FakeWebSocket.last!.close).toHaveBeenCalled();
  });

  it("ignores malformed frames", async () => {
    const events: ScanRecorded[] = [];
    openScanSocket(
      async () => "tok",
      (e) => events.push(e),
    );

    await vi.waitFor(() => expect(FakeWebSocket.last).not.toBeNull());
    FakeWebSocket.last!.onmessage?.({ data: "{not json" });
    expect(events).toHaveLength(0);
  });
});

describe("isTaskNotification", () => {
  it("accepts assigned and deadline frames", async () => {
    const { isTaskNotification } = await import("../ws");
    expect(isTaskNotification({ task_id: "t1", title: "Restock", kind: "assigned" })).toBe(true);
    expect(isTaskNotification({ task_id: "t1", title: "Restock", kind: "deadline" })).toBe(true);
  });

  it("rejects non-task frames", async () => {
    const { isTaskNotification } = await import("../ws");
    expect(isTaskNotification({ sku: "X", action: "place" })).toBe(false);
    expect(isTaskNotification({ task_id: "t1", title: "x", kind: "other" })).toBe(false);
    expect(isTaskNotification(null)).toBe(false);
  });
});

describe("openTaskNotificationSocket", () => {
  it("forwards only task-notification frames", async () => {
    const { openTaskNotificationSocket } = await import("../ws");
    const got: unknown[] = [];
    const close = openTaskNotificationSocket(
      async () => "tok",
      (n) => got.push(n),
    );

    await vi.waitFor(() => expect(FakeWebSocket.last).not.toBeNull());
    // a scan frame must be ignored
    FakeWebSocket.last!.onmessage?.({ data: JSON.stringify({ sku: "X", action: "place" }) });
    // a task notification must pass
    FakeWebSocket.last!.onmessage?.({
      data: JSON.stringify({ task_id: "t1", title: "Restock", kind: "assigned", ts: "t" }),
    });

    expect(got).toHaveLength(1);
    close();
  });
});
