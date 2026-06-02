import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { useCompleteCount, useSetCountLine, useStartCount } from "../useCounts";

const mockPost = vi.fn();
const mockPatch = vi.fn();

vi.mock("../../../lib/api", () => ({
  useApi: () => ({ post: mockPost, patch: mockPatch }),
}));
vi.mock("react-oidc-context", () => ({ useAuth: () => ({ user: { profile: {} } }) }));

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return createElement(QueryClientProvider, { client: qc }, children);
}

describe("useCounts", () => {
  it("starts a count with the zone id", async () => {
    mockPost.mockResolvedValue({ id: "c1", zone_id: "z1", status: "open", lines: [] });
    const { result } = renderHook(() => useStartCount("store-1"), { wrapper });
    result.current.mutate("z1");
    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith("/api/v1/stores/store-1/counts", { zone_id: "z1" }),
    );
  });

  it("patches a counted line", async () => {
    mockPatch.mockResolvedValue({});
    const { result } = renderHook(() => useSetCountLine("store-1", "c1"), { wrapper });
    result.current.mutate({ sku: "SKU-1", counted_qty: 5 });
    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith("/api/v1/stores/store-1/counts/c1/lines", {
        sku: "SKU-1",
        counted_qty: 5,
      }),
    );
  });

  it("completes a count", async () => {
    mockPost.mockResolvedValue({ id: "c1", status: "completed", reconciled: 0, variances: [] });
    const { result } = renderHook(() => useCompleteCount("store-1", "c1"), { wrapper });
    result.current.mutate();
    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith("/api/v1/stores/store-1/counts/c1/complete", {}),
    );
  });
});
