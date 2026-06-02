import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ToastProvider } from "../../../components/feedback/toast";
import { TransferStockModal } from "../TransferStockModal";
import type { InventoryRow } from "../types";

const mockPost = vi.fn();
const mockGet = vi.fn();

vi.mock("../../../lib/api", () => ({ useApi: () => ({ post: mockPost, get: mockGet }) }));
vi.mock("react-oidc-context", () => ({ useAuth: () => ({ user: { profile: {} } }) }));
vi.mock("../../map/useCurrentStore", () => ({ useCurrentStore: () => "store-1" }));

const ZONES = [
  { id: "zone-1", name: "Frozen", type: "general", x: 0, y: 0, width: 100, height: 100 },
  { id: "zone-2", name: "Backroom", type: "general", x: 0, y: 0, width: 100, height: 100 },
];

const ROW: InventoryRow = {
  id: "loc-1",
  zone_id: "zone-1",
  sku: "SKU-7",
  name: "Widget",
  category: "frozen",
  status: "active",
  qty: 10,
  updated_at: "t0",
};

function renderModal(onClose = vi.fn()) {
  mockGet.mockResolvedValue(ZONES);
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <TransferStockModal row={ROW} onClose={onClose} />
      </ToastProvider>
    </QueryClientProvider>,
  );
}

describe("TransferStockModal", () => {
  it("posts the transfer payload and closes", async () => {
    const onClose = vi.fn();
    mockPost.mockResolvedValue({});
    renderModal(onClose);

    // pick destination zone via custom Select
    fireEvent.click(screen.getByText("Select destination…"));
    fireEvent.click(await screen.findByRole("option", { name: /Backroom/i }));

    fireEvent.change(screen.getByDisplayValue("1"), { target: { value: "4" } });
    fireEvent.click(screen.getByRole("button", { name: "Move stock" }));

    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith("/api/v1/stores/store-1/inventory/transfer", {
        sku: "SKU-7",
        from_zone_id: "zone-1",
        to_zone_id: "zone-2",
        qty: 4,
      }),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it("disables submit until a destination is chosen", () => {
    renderModal();
    expect(screen.getByRole("button", { name: "Move stock" })).toBeDisabled();
  });
});
