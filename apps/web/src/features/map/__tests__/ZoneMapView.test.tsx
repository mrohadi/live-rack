import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ZoneMapView } from "../ZoneMapView";
import { fillRatio, slotStatus } from "../zoneMath";
import type { Zone } from "../types";
import type { InventoryRow } from "../../inventory/types";

const zones: Zone[] = [
  {
    id: "A2",
    name: "A2 · Apparel",
    x: 30,
    y: 4,
    width: 28,
    height: 28,
    color: "#2563eb",
    type: "general",
    items: 412,
    capacity: 480,
  },
];

const items: InventoryRow[] = [
  {
    id: "i1",
    zone_id: "A2",
    sku: "LR-7821",
    name: "Sweater",
    category: "apparel",
    status: "active",
    qty: 24,
    updated_at: "2026-06-02T00:00:00Z",
  },
];

describe("fillRatio", () => {
  it("clamps to capacity", () => {
    expect(fillRatio(zones[0])).toBeCloseTo(412 / 480);
    expect(fillRatio({ ...zones[0], items: 999 })).toBe(1);
    expect(fillRatio({ ...zones[0], capacity: 0 })).toBe(0);
  });
});

describe("slotStatus", () => {
  it("bands by quantity", () => {
    expect(slotStatus(0)).toBe("out");
    expect(slotStatus(5)).toBe("low");
    expect(slotStatus(40)).toBe("ok");
  });
});

describe("ZoneMapView", () => {
  it("renders a zone box with fill meta", () => {
    render(
      <ZoneMapView zones={zones} items={items} view="zones" selectedId="A2" onSelect={() => {}} />,
    );
    expect(screen.getByText("A2 · Apparel")).toBeInTheDocument();
    expect(screen.getByText(/412\/480 · 86%/)).toBeInTheDocument();
  });

  it("selects a zone on click", async () => {
    const onSelect = vi.fn();
    render(
      <ZoneMapView
        zones={zones}
        items={items}
        view="zones"
        selectedId={null}
        onSelect={onSelect}
      />,
    );
    screen.getByTestId("zone-box").click();
    expect(onSelect).toHaveBeenCalledWith("A2");
  });
});
