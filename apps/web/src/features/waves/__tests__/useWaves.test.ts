import { describe, expect, it } from "vitest";
import type { WaveStop } from "../types";
import { stopToLine, waveProgress } from "../useWaves";

function stop(seq: number, status: WaveStop["status"]): WaveStop {
  return {
    seq,
    sku: `SKU-${seq}`,
    zone_id: `z${seq}`,
    zone_name: "A1",
    zone_x: seq,
    zone_y: 0,
    qty_requested: 6,
    qty_picked: status === "pending" ? 0 : 6,
    order_count: 2,
    status,
  };
}

describe("waveProgress", () => {
  it("counts non-pending stops and percent", () => {
    const stops = [stop(0, "picked"), stop(1, "short"), stop(2, "pending")];
    expect(waveProgress(stops)).toEqual({ done: 2, total: 3, pct: 67 });
  });

  it("is zero for no stops", () => {
    expect(waveProgress([])).toEqual({ done: 0, total: 0, pct: 0 });
  });
});

describe("stopToLine", () => {
  it("maps a stop to a route line with a composite id", () => {
    const line = stopToLine(stop(1, "pending"));
    expect(line.id).toBe("SKU-1|z1");
    expect(line.seq).toBe(1);
    expect(line.qty_requested).toBe(6);
    expect(line.status).toBe("pending");
  });
});
