import { render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { KonvaZoneCanvas } from "../renderers/KonvaZoneCanvas";
import type { Zone } from "../types";

const zones: Zone[] = [
  { id: "a", name: "A", x: 0, y: 0, width: 100, height: 100, color: "#6366f1", type: "general" },
  { id: "b", name: "B", x: 200, y: 0, width: 100, height: 100, color: "#10b981", type: "frozen" },
];

describe("KonvaZoneCanvas — editor mode", () => {
  it("renders as editable when onChange provided", () => {
    const { container } = render(
      <KonvaZoneCanvas zones={zones} selectedIds={[]} onSelect={vi.fn()} onChange={vi.fn()} />,
    );
    expect(container.querySelector("canvas")).toBeTruthy();
  });

  it("renders read-only without onChange", () => {
    const { container } = render(
      <KonvaZoneCanvas zones={zones} selectedIds={[]} onSelect={vi.fn()} />,
    );
    expect(container.querySelector("canvas")).toBeTruthy();
  });

  it("accepts gridSize prop", () => {
    expect(() =>
      render(
        <KonvaZoneCanvas
          zones={zones}
          selectedIds={[]}
          onSelect={vi.fn()}
          onChange={vi.fn()}
          gridSize={25}
        />,
      ),
    ).not.toThrow();
  });
});
