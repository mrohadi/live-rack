import { describe, expect, it } from "vitest";
import { shipmentStatusLabel } from "../useShipments";

describe("shipmentStatusLabel", () => {
  it("maps each status to a label", () => {
    expect(shipmentStatusLabel("packing")).toBe("Packing");
    expect(shipmentStatusLabel("packed")).toBe("Packed");
    expect(shipmentStatusLabel("dispatched")).toBe("Dispatched");
    expect(shipmentStatusLabel("cancelled")).toBe("Cancelled");
  });
});
