import { describe, expect, it } from "vitest";
import { formatCents, sparkPoints } from "../useSales";

describe("formatCents", () => {
  it("renders cents as USD currency", () => {
    expect(formatCents(12345)).toBe("$123.45");
    expect(formatCents(0)).toBe("$0.00");
  });
});

describe("sparkPoints", () => {
  it("maps values into scaled svg coordinates", () => {
    const pts = sparkPoints([0, 5, 10], 100, 40).split(" ");
    expect(pts).toHaveLength(3);
    expect(pts[0]).toBe("0.0,40.0"); // min → bottom
    expect(pts[2]).toBe("100.0,0.0"); // max → top, last x = width
  });

  it("returns empty for no data", () => {
    expect(sparkPoints([], 100, 40)).toBe("");
  });
});
