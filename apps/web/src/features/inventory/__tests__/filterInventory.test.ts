import { describe, expect, it } from "vitest";
import type { InventoryRow } from "../types";
import { filterInventory, rowStockStatus, rowVelocity } from "../useInventory";

function row(over: Partial<InventoryRow>): InventoryRow {
  return {
    id: "r1",
    zone_id: "z1",
    sku: "SKU-1",
    name: "Widget",
    category: "frozen",
    status: "active",
    qty: 5,
    updated_at: "t0",
    ...over,
  };
}

const rows: InventoryRow[] = [
  row({ id: "a", zone_id: "z1", status: "active", velocity: "hot" }),
  row({ id: "b", zone_id: "z2", status: "discontinued", velocity: "cold" }),
  row({ id: "c", zone_id: "z1", status: "recalled" }), // velocity undefined → cold
];

const ALL = { zone: "all", status: "all", velocity: "all", stock: "all" };

describe("rowVelocity", () => {
  it("defaults missing velocity to cold", () => {
    expect(rowVelocity(row({ velocity: undefined }))).toBe("cold");
    expect(rowVelocity(row({ velocity: "hot" }))).toBe("hot");
  });
});

describe("rowStockStatus", () => {
  it("out when qty is zero", () => {
    expect(rowStockStatus(row({ qty: 0 }))).toBe("out");
  });
  it("low when qty at/below reorder point", () => {
    expect(rowStockStatus(row({ qty: 3, reorder_point: 5 }))).toBe("low");
  });
  it("in_stock above reorder point", () => {
    expect(rowStockStatus(row({ qty: 9, reorder_point: 5 }))).toBe("in_stock");
  });
  it("in_stock when reorder point disabled and positive", () => {
    expect(rowStockStatus(row({ qty: 1, reorder_point: 0 }))).toBe("in_stock");
  });
  it("prefers server-provided stock_status", () => {
    expect(rowStockStatus(row({ qty: 100, stock_status: "low" }))).toBe("low");
  });
});

describe("filterInventory", () => {
  it("returns all rows when every filter is 'all'", () => {
    expect(filterInventory(rows, ALL)).toHaveLength(3);
  });

  it("filters by zone", () => {
    expect(filterInventory(rows, { ...ALL, zone: "z2" }).map((r) => r.id)).toEqual(["b"]);
  });

  it("filters by status", () => {
    expect(filterInventory(rows, { ...ALL, status: "recalled" }).map((r) => r.id)).toEqual(["c"]);
  });

  it("filters by velocity, treating undefined as cold", () => {
    expect(filterInventory(rows, { ...ALL, velocity: "cold" }).map((r) => r.id)).toEqual([
      "b",
      "c",
    ]);
  });

  it("combines filters (AND semantics)", () => {
    expect(
      filterInventory(rows, { zone: "z1", status: "active", velocity: "hot" }).map((r) => r.id),
    ).toEqual(["a"]);
  });
});
