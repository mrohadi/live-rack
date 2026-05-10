/**
 * LR-101 spike: Konva vs SVG renderer for 500+ zones
 *
 * Performance criteria (JSDOM approximations — real perf measured manually):
 *   - Must mount 500 zones without throwing
 *   - Must expose each zone by its id for event wiring
 *   - Must update a single zone without remounting all siblings (key/id stability)
 *   - Re-render with changed selection must not error
 */
import { cleanup, fireEvent, render } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { KonvaZoneCanvas } from "../renderers/KonvaZoneCanvas";
import { SvgZoneCanvas } from "../renderers/SvgZoneCanvas";
import type { Zone } from "../types";

afterEach(cleanup);

function makeZones(n: number): Zone[] {
  return Array.from({ length: n }, (_, i) => ({
    id: `zone-${i}`,
    name: `Zone ${i}`,
    x: (i % 25) * 44,
    y: Math.floor(i / 25) * 44,
    width: 40,
    height: 40,
    color: "#6366f1",
    type: "general" as const,
  }));
}

const ZONE_COUNT = 500;
const zones = makeZones(ZONE_COUNT);

describe("SvgZoneCanvas", () => {
  it("renders 500 zones without throwing", () => {
    const { container } = render(
      <SvgZoneCanvas zones={zones} selectedId={null} onSelect={() => {}} />,
    );
    const rects = container.querySelectorAll("[data-zone-id]");
    expect(rects).toHaveLength(ZONE_COUNT);
  });

  it("marks the selected zone", () => {
    render(<SvgZoneCanvas zones={zones} selectedId="zone-42" onSelect={() => {}} />);
    const selected = document.querySelector('[data-zone-id="zone-42"][data-selected="true"]');
    expect(selected).not.toBeNull();
  });

  it("calls onSelect with zone id on click", async () => {
    const onSelect = vi.fn();
    const { container } = render(
      <SvgZoneCanvas zones={zones} selectedId={null} onSelect={onSelect} />,
    );
    const first = container.querySelector('[data-zone-id="zone-0"]') as Element;
    fireEvent.click(first);
    expect(onSelect).toHaveBeenCalledWith("zone-0");
  });

  it("re-renders changed selection without mounting new zones", () => {
    const { rerender, container } = render(
      <SvgZoneCanvas zones={zones} selectedId={null} onSelect={() => {}} />,
    );
    const beforeCount = container.querySelectorAll("[data-zone-id]").length;
    rerender(<SvgZoneCanvas zones={zones} selectedId="zone-10" onSelect={() => {}} />);
    const afterCount = container.querySelectorAll("[data-zone-id]").length;
    expect(afterCount).toBe(beforeCount);
  });
});

describe("KonvaZoneCanvas", () => {
  it("renders 500 zones without throwing", () => {
    const { container } = render(
      <KonvaZoneCanvas zones={zones} selectedId={null} onSelect={() => {}} />,
    );
    expect(container.querySelector("canvas")).not.toBeNull();
  });

  it("accepts selectedId and onSelect props", () => {
    const onSelect = vi.fn();
    expect(() =>
      render(<KonvaZoneCanvas zones={zones} selectedId="zone-42" onSelect={onSelect} />),
    ).not.toThrow();
  });

  it("re-renders selection change without throwing", () => {
    const { rerender } = render(
      <KonvaZoneCanvas zones={zones} selectedId={null} onSelect={() => {}} />,
    );
    expect(() =>
      rerender(<KonvaZoneCanvas zones={zones} selectedId="zone-10" onSelect={() => {}} />),
    ).not.toThrow();
  });
});
