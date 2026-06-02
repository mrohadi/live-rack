import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ToastProvider } from "../../../components/feedback/toast";
import { ItemDetailDrawer } from "../ItemDetailDrawer";
import type { ItemDetail } from "../types";

const mockGet = vi.fn();
const mockPatch = vi.fn();

vi.mock("../../../lib/api", () => ({ useApi: () => ({ get: mockGet, patch: mockPatch }) }));
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
      <ToastProvider>
        <ItemDetailDrawer sku="SKU-7" onClose={onClose} />
      </ToastProvider>
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

  it("edits catalog fields via PATCH /inventory/:sku", async () => {
    mockPatch.mockResolvedValue({});
    renderDrawer();
    await screen.findByText("Widget");

    fireEvent.click(screen.getByRole("button", { name: "Edit" }));
    fireEvent.change(screen.getByDisplayValue("Widget"), { target: { value: "Widget Pro" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith(
        "/api/v1/stores/store-1/inventory/SKU-7",
        expect.objectContaining({ name: "Widget Pro", reorder_point: 5 }),
      ),
    );
  });

  it("corrects a zone qty via PATCH /inventory/:sku/qty", async () => {
    mockPatch.mockResolvedValue({});
    renderDrawer();
    await screen.findByText("Widget");

    // zone-1 (Frozen) shows qty 3 as a button; click to edit
    fireEvent.click(screen.getByRole("button", { name: "3" }));
    fireEvent.change(screen.getByDisplayValue("3"), { target: { value: "2" } });
    fireEvent.click(screen.getByRole("button", { name: "✓" }));

    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith("/api/v1/stores/store-1/inventory/SKU-7/qty", {
        zone_id: "z1",
        qty: 2,
      }),
    );
  });
});
