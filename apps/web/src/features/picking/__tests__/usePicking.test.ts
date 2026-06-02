import { describe, expect, it } from "vitest";
import type { PickLine } from "../types";
import { lineStatusLabel, nextPendingLine, pickProgress } from "../usePicking";

function line(seq: number, status: PickLine["status"]): PickLine {
  return {
    id: `l${seq}`,
    seq,
    sku: `SKU-${seq}`,
    zone_name: "A1",
    zone_x: 0,
    zone_y: 0,
    qty_requested: 5,
    qty_picked: status === "pending" ? 0 : 5,
    status,
  };
}

describe("pickProgress", () => {
  it("counts non-pending lines and computes percent", () => {
    const lines = [line(0, "picked"), line(1, "short"), line(2, "pending")];
    expect(pickProgress(lines)).toEqual({ done: 2, total: 3, pct: 67 });
  });

  it("is zero for an empty route", () => {
    expect(pickProgress([])).toEqual({ done: 0, total: 0, pct: 0 });
  });
});

describe("nextPendingLine", () => {
  it("returns the lowest-seq pending stop", () => {
    const lines = [line(2, "pending"), line(0, "picked"), line(1, "pending")];
    expect(nextPendingLine(lines)?.seq).toBe(1);
  });

  it("returns undefined when all picked", () => {
    expect(nextPendingLine([line(0, "picked"), line(1, "short")])).toBeUndefined();
  });
});

describe("lineStatusLabel", () => {
  it("maps statuses to labels", () => {
    expect(lineStatusLabel("picked")).toBe("Picked");
    expect(lineStatusLabel("short")).toBe("Short");
    expect(lineStatusLabel("pending")).toBe("Pending");
  });
});
