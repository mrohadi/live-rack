import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ZonePerfBars } from "../ZonePerfBars";
import { barWidthPct, maxScans, type ZonePerf } from "../useAnalytics";

const zones: ZonePerf[] = [
  { zone_id: "aaaaaaaa-1111-2222-3333-444444444444", scans: 30, picks: 20, invalid: 2, spark: [10, 20] },
  { zone_id: "bbbbbbbb-1111-2222-3333-444444444444", scans: 15, picks: 5, invalid: 0, spark: [5] },
];

describe("maxScans / barWidthPct", () => {
  it("computes the peak and relative widths", () => {
    expect(maxScans(zones)).toBe(30);
    expect(barWidthPct(30, zones)).toBe(100);
    expect(barWidthPct(15, zones)).toBe(50);
  });

  it("floors max at 1 for empty/zero data", () => {
    expect(maxScans([])).toBe(1);
  });
});

describe("ZonePerfBars", () => {
  it("renders a row per zone and flags invalid scans", () => {
    render(<ZonePerfBars zones={zones} />);
    expect(screen.getByText("aaaaaaaa")).toBeInTheDocument();
    expect(screen.getByText("⚠ 2")).toBeInTheDocument();
    expect(screen.getAllByText("30")).toHaveLength(1);
  });

  it("shows an empty state with no zones", () => {
    render(<ZonePerfBars zones={[]} />);
    expect(screen.getByText(/no zone activity/i)).toBeInTheDocument();
  });
});
