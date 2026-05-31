import { describe, expect, it } from "vitest";
import { cellIntensity, heatColor } from "../useAnalytics";

describe("cellIntensity", () => {
  it("normalises against the peak", () => {
    expect(cellIntensity(5, 10)).toBe(0.5);
    expect(cellIntensity(10, 10)).toBe(1);
    expect(cellIntensity(0, 10)).toBe(0);
  });

  it("guards zero/negative max", () => {
    expect(cellIntensity(3, 0)).toBe(0);
  });

  it("clamps to 0..1", () => {
    expect(cellIntensity(20, 10)).toBe(1);
    expect(cellIntensity(-5, 10)).toBe(0);
  });
});

describe("heatColor", () => {
  it("blends accent over panel by intensity percent", () => {
    expect(heatColor(5, 10, "var(--accent)")).toBe(
      "color-mix(in oklab, var(--accent) 50%, var(--panel))",
    );
    expect(heatColor(0, 10, "#fff")).toBe("color-mix(in oklab, #fff 0%, var(--panel))");
  });
});
