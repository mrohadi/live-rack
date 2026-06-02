import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ToastProvider } from "../../../components/feedback/toast";
import { AddItemModal } from "../AddItemModal";

const mockPost = vi.fn();
const mockGet = vi.fn();

vi.mock("../../../lib/api", () => ({ useApi: () => ({ post: mockPost, get: mockGet }) }));
vi.mock("react-oidc-context", () => ({
  useAuth: () => ({ user: { profile: {} } }),
}));
vi.mock("../../map/useCurrentStore", () => ({ useCurrentStore: () => "store-1" }));

const ZONES = [
  { id: "zone-1", name: "Zone Alpha", type: "general", x: 0, y: 0, width: 100, height: 100 },
];

function renderModal(props: { defaultZoneId?: string; onClose?: () => void } = {}) {
  mockGet.mockResolvedValue(ZONES);
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <AddItemModal onClose={props.onClose ?? vi.fn()} defaultZoneId={props.defaultZoneId} />
      </ToastProvider>
    </QueryClientProvider>,
  );
}

describe("AddItemModal", () => {
  it("renders form fields", () => {
    renderModal();
    expect(screen.getByRole("dialog", { name: "Add item to zone" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. SKU-1234")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. Widget Blue")).toBeInTheDocument();
  });

  it("hides zone selector when defaultZoneId provided", () => {
    renderModal({ defaultZoneId: "zone-1" });
    expect(screen.queryByText("Select zone…")).not.toBeInTheDocument();
  });

  it("submits with correct payload and calls onClose", async () => {
    const onClose = vi.fn();
    mockPost.mockResolvedValue({
      id: "loc-1",
      zone_id: "zone-1",
      sku: "SKU-42",
      name: "Test Widget",
      category: "general",
      status: "active",
      qty: 3,
      updated_at: new Date().toISOString(),
      velocity: "cold",
    });

    renderModal({ defaultZoneId: "zone-1", onClose });

    fireEvent.change(screen.getByPlaceholderText("e.g. SKU-1234"), {
      target: { value: "SKU-42" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. Widget Blue"), {
      target: { value: "Test Widget" },
    });
    fireEvent.change(screen.getByPlaceholderText("e.g. frozen"), {
      target: { value: "general" },
    });
    // qty input — find by type=number
    const qtyInput = screen.getByDisplayValue("1");
    fireEvent.change(qtyInput, { target: { value: "3" } });

    fireEvent.click(screen.getByRole("button", { name: "Add item" }));

    await waitFor(() => expect(mockPost).toHaveBeenCalledWith(
      "/api/v1/stores/store-1/inventory",
      expect.objectContaining({ sku: "SKU-42", qty: 3, zone_id: "zone-1" }),
    ));
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it("calls onClose when Cancel clicked", () => {
    const onClose = vi.fn();
    renderModal({ onClose });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalled();
  });
});
