import { describe, expect, it } from "vitest";
import type { PickLine } from "../types";
import { nextScanQty, normalizeScan, resolveScan } from "../pickScan";

function line(seq: number, sku: string, status: PickLine["status"]): PickLine {
  return {
    id: `l${seq}`,
    seq,
    sku,
    zone_name: "A1",
    zone_x: 0,
    zone_y: 0,
    qty_requested: 3,
    qty_picked: status === "pending" ? 0 : 3,
    status,
  };
}

describe("normalizeScan", () => {
  it("trims and upper-cases", () => {
    expect(normalizeScan(" sku-1 ")).toBe("SKU-1");
  });
});

describe("resolveScan", () => {
  const lines = [
    line(0, "SKU-A", "picked"),
    line(1, "SKU-B", "pending"),
    line(2, "SKU-A", "pending"),
  ];

  it("credits the lowest-seq pending line for the SKU", () => {
    expect(resolveScan("sku-a", lines)?.id).toBe("l2");
  });

  it("matches case-insensitively and trims", () => {
    expect(resolveScan("  sku-b ", lines)?.id).toBe("l1");
  });

  it("returns undefined on a mis-scan", () => {
    expect(resolveScan("SKU-Z", lines)).toBeUndefined();
  });
});

describe("nextScanQty", () => {
  it("increments toward and caps at requested", () => {
    expect(nextScanQty(0, 3)).toBe(1);
    expect(nextScanQty(2, 3)).toBe(3);
    expect(nextScanQty(3, 3)).toBe(3);
  });
});
