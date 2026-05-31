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
