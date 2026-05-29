import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ZoneDetailSidebar } from "../ZoneDetailSidebar";
import type { Zone } from "../types";

const zone: Zone = {
  id: "z1",
  name: "Zone A",
  x: 0,
  y: 0,
  width: 200,
  height: 140,
  color: "#6366f1",
  type: "general",
  items: 140,
  capacity: 200,
  sales: 4320,
  misplaced: 3,
  lastScan: "14:23:08",
  dwell: "2d 4h",
  constraints: {
    allowedCategories: ["Apparel"],
    maxSKUs: 480,
    climate: "ambient",
  },
};

describe("ZoneDetailSidebar", () => {
  it("renders zone name and id", () => {
    render(<ZoneDetailSidebar zone={zone} />);
    expect(screen.getByText("Zone A")).toBeInTheDocument();
    expect(screen.getByText(/zone z1/i)).toBeInTheDocument();
  });

  it("renders fill percentage", () => {
    render(<ZoneDetailSidebar zone={zone} />);
    expect(screen.getByText("70%")).toBeInTheDocument();
  });

  it("renders capacity kv", () => {
    render(<ZoneDetailSidebar zone={zone} />);
    expect(screen.getByText("140 / 200")).toBeInTheDocument();
  });

  it("renders constraint chips", () => {
    render(<ZoneDetailSidebar zone={zone} />);
    expect(screen.getByText(/Apparel/i)).toBeInTheDocument();
    expect(screen.getByText(/Max 480 SKUs/i)).toBeInTheDocument();
    expect(screen.getByText(/ambient/i)).toBeInTheDocument();
  });

  it("renders empty state when no zone", () => {
    render(<ZoneDetailSidebar zone={null} />);
    expect(screen.getByText(/select a zone/i)).toBeInTheDocument();
  });

  it("renders action buttons", () => {
    render(<ZoneDetailSidebar zone={zone} />);
    expect(screen.getByRole("button", { name: /open zone/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /assign task/i })).toBeInTheDocument();
  });
});
