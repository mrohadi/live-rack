import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Heatmap } from "../Heatmap";
import { HOURS_IN_DAY, type HeatmapResponse } from "../useAnalytics";

function emptyGrid(): HeatmapResponse {
  const grid = Array.from({ length: 7 }, () => Array.from({ length: HOURS_IN_DAY }, () => 0));
  return { grid, max: 0 };
}

describe("Heatmap", () => {
  it("renders 7 day rows with 24 cells each", () => {
    const data = emptyGrid();
    data.grid[0][9] = 5; // Mon 09:00
    data.max = 5;

    render(<Heatmap data={data} />);

    expect(screen.getByText("Mon")).toBeInTheDocument();
    expect(screen.getByText("Sun")).toBeInTheDocument();
    // 7 rows * 24 cells, each carries a title tooltip.
    expect(screen.getByTitle("Mon 9:00 · 5")).toBeInTheDocument();
    expect(screen.getAllByTitle(/·/)).toHaveLength(7 * HOURS_IN_DAY);
  });
});
