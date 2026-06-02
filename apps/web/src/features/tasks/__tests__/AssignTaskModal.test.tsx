import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { beforeEach, describe, it, expect, vi } from "vitest";
import { ToastProvider } from "../../../components/feedback/toast";
import { AssignTaskModal } from "../AssignTaskModal";

const mockPost = vi.fn();

vi.mock("../../../lib/api", () => ({ useApi: () => ({ post: mockPost }) }));
vi.mock("react-oidc-context", () => ({
  useAuth: () => ({ user: { profile: {} } }),
}));
vi.mock("../../map/useCurrentStore", () => ({ useCurrentStore: () => "store-1" }));

function renderModal(props: { zoneId?: string; zoneName?: string; onClose?: () => void } = {}) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <AssignTaskModal
          zoneId={props.zoneId ?? "zone-1"}
          zoneName={props.zoneName ?? "Frozen"}
          onClose={props.onClose ?? vi.fn()}
        />
      </ToastProvider>
    </QueryClientProvider>,
  );
}

describe("AssignTaskModal", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders dialog with zone name", () => {
    renderModal({ zoneName: "Frozen" });
    expect(screen.getByRole("dialog", { name: "Assign task" })).toBeInTheDocument();
    expect(screen.getByText("Frozen")).toBeInTheDocument();
  });

  it("submits with correct payload and calls onClose", async () => {
    const onClose = vi.fn();
    mockPost.mockResolvedValue({
      id: "task-1",
      store_id: "store-1",
      zone_id: "zone-1",
      title: "Restock frozen",
      status: "todo",
      priority: "high",
      updated_at: new Date().toISOString(),
    });

    renderModal({ zoneId: "zone-1", zoneName: "Frozen", onClose });

    fireEvent.change(screen.getByPlaceholderText("e.g. Restock frozen section"), {
      target: { value: "Restock frozen" },
    });

    // Custom Select — open priority dropdown then click "High" option.
    fireEvent.click(screen.getByText("Medium"));
    fireEvent.click(screen.getByRole("option", { name: /high/i }));

    fireEvent.click(screen.getByRole("button", { name: "Create task" }));

    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith(
        "/api/v1/stores/store-1/tasks",
        expect.objectContaining({
          zone_id: "zone-1",
          title: "Restock frozen",
          priority: "high",
        }),
      ),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it("calls onClose when Cancel clicked", () => {
    const onClose = vi.fn();
    renderModal({ onClose });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalled();
  });

  it("does not submit when title is empty", () => {
    renderModal();
    fireEvent.click(screen.getByRole("button", { name: "Create task" }));
    expect(mockPost).not.toHaveBeenCalled();
  });
});
