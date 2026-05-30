import { afterEach, describe, expect, it } from "vitest";
import { closeQueue, enqueueScan, flushQueue, pendingCount, type ScanPayload } from "../scanQueue";

const scan: ScanPayload = { sku: "ABC-1", zoneId: "z1", scannedAt: 1 };

afterEach(async () => {
  await closeQueue();
  await new Promise<void>((resolve) => {
    const req = indexedDB.deleteDatabase("live-rack-scanner");
    req.onsuccess = req.onerror = req.onblocked = () => resolve();
  });
});

describe("scanQueue", () => {
  it("enqueues a scan and counts it", async () => {
    await enqueueScan(scan);
    expect(await pendingCount()).toBe(1);
  });

  it("flushes queued scans and clears them on success", async () => {
    await enqueueScan(scan);
    await enqueueScan({ ...scan, sku: "ABC-2" });
    const sent: ScanPayload[] = [];
    await flushQueue(async (s) => {
      sent.push(s);
    });
    expect(sent).toHaveLength(2);
    expect(await pendingCount()).toBe(0);
  });

  it("stops on first failure and keeps remaining scans", async () => {
    await enqueueScan(scan);
    await enqueueScan({ ...scan, sku: "ABC-2" });
    await expect(
      flushQueue(async () => {
        throw new Error("offline");
      }),
    ).rejects.toThrow("offline");
    expect(await pendingCount()).toBe(2);
  });
});
