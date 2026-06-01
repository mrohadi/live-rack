import { describe, expect, it } from "vitest";
import { findOpenSlot, rectsOverlap } from "../zoneMath";

describe("rectsOverlap", () => {
  const a = { x: 0, y: 0, width: 20, height: 20 };
  it("detects overlap", () => {
    expect(rectsOverlap(a, { x: 10, y: 10, width: 20, height: 20 })).toBe(true);
  });
  it("detects separation", () => {
    expect(rectsOverlap(a, { x: 30, y: 0, width: 20, height: 20 })).toBe(false);
  });
  it("honors the gap", () => {
    const b = { x: 22, y: 0, width: 10, height: 10 };
    expect(rectsOverlap(a, b)).toBe(false);
    expect(rectsOverlap(a, b, 4)).toBe(true);
  });
});

describe("findOpenSlot", () => {
  it("returns origin on an empty canvas", () => {
    expect(findOpenSlot([], 20, 14)).toEqual({ x: 0, y: 0 });
  });

  it("avoids an existing zone at the origin", () => {
    const slot = findOpenSlot([{ x: 0, y: 0, width: 30, height: 30 }], 20, 14);
    const candidate = { ...slot, width: 20, height: 14 };
    expect(rectsOverlap(candidate, { x: 0, y: 0, width: 30, height: 30 }, 2)).toBe(false);
  });
});
