import { describe, expect, it } from "vitest";
import type { ScanRecorded } from "../../../lib/ws";
import type { InventoryRow } from "../types";
import { patchInventory, scanQtyDelta } from "../useInventory";

const baseRow: InventoryRow = {
  id: "r1",
  zone_id: "z1",
  sku: "SKU-1",
  name: "Widget",
  category: "frozen",
  status: "active",
  qty: 5,
  updated_at: "t0",
};

function scan(over: Partial<ScanRecorded>): ScanRecorded {
  return {
    org_id: "o1",
    store_id: "s1",
    zone_id: "z1",
    scanner_id: "scn",
    sku: "SKU-1",
    action: "place",
    valid: true,
    ts: "t1",
    ...over,
  };
}

describe("scanQtyDelta", () => {
  it("removes stock on pick, adds otherwise", () => {
    expect(scanQtyDelta("pick")).toBe(-1);
    expect(scanQtyDelta("place")).toBe(1);
    expect(scanQtyDelta("count")).toBe(1);
  });
});

describe("patchInventory", () => {
  it("increments the matching row on a valid place", () => {
    const next = patchInventory([baseRow], scan({ action: "place" }));
    expect(next?.[0].qty).toBe(6);
    expect(next?.[0].updated_at).toBe("t1");
  });

  it("decrements but floors at zero on pick", () => {
    const next = patchInventory([{ ...baseRow, qty: 0 }], scan({ action: "pick" }));
    expect(next?.[0].qty).toBe(0);
  });

  it("ignores invalid scans", () => {
    const next = patchInventory([baseRow], scan({ valid: false }));
    expect(next?.[0].qty).toBe(5);
  });

  it("leaves non-matching rows untouched", () => {
    const next = patchInventory([baseRow], scan({ sku: "OTHER" }));
    expect(next?.[0].qty).toBe(5);
  });
});
