import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ItemDetailDrawer } from "../ItemDetailDrawer";
import type { ItemDetail } from "../types";

const mockGet = vi.fn();

vi.mock("../../../lib/api", () => ({ useApi: () => ({ get: mockGet }) }));
vi.mock("react-oidc-context", () => ({ useAuth: () => ({ user: { profile: {} } }) }));
vi.mock("../../map/useCurrentStore", () => ({ useCurrentStore: () => "store-1" }));

const DETAIL: ItemDetail = {
  sku: "SKU-7",
  name: "Widget",
  category: "frozen",
  status: "active",
  reorder_point: 5,
  total_qty: 12,
  stock_status: "in_stock",
  locations: [
    { zone_id: "z1", zone_name: "Frozen", qty: 3, stock_status: "low", updated_at: "t0" },
    { zone_id: "z2", zone_name: "Backroom", qty: 9, stock_status: "in_stock", updated_at: "t0" },
  ],
  recent_scans: [
    {
      ts: new Date().toISOString(),
      zone_id: "z1",
      scanner_id: "scn-1",
      action: "pick",
      valid: true,
    },
  ],
};

function renderDrawer(onClose = vi.fn()) {
  mockGet.mockResolvedValue(DETAIL);
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <ItemDetailDrawer sku="SKU-7" onClose={onClose} />
    </QueryClientProvider>,
  );
}

describe("ItemDetailDrawer", () => {
  it("requests the item and renders zones + scans", async () => {
    renderDrawer();
    await waitFor(() =>
      expect(mockGet).toHaveBeenCalledWith("/api/v1/stores/store-1/inventory/SKU-7"),
    );
    expect(await screen.findByText("Widget")).toBeInTheDocument();
    expect(screen.getByText("Frozen")).toBeInTheDocument();
    expect(screen.getByText("Backroom")).toBeInTheDocument();
    expect(screen.getByText("pick")).toBeInTheDocument();
  });

  it("closes via the close button", async () => {
    const onClose = vi.fn();
    renderDrawer(onClose);
    await screen.findByText("Widget");
    fireEvent.click(screen.getByLabelText("Close"));
    expect(onClose).toHaveBeenCalled();
  });
});
