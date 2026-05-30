import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { Zone } from "../types";
import { useCreateZone, useUpdateZone, useZones } from "../useZones";

// Mock the API client — we assert on the calls the hooks make, not on fetch/Clerk.
const get = vi.fn();
const post = vi.fn();
const put = vi.fn();
vi.mock("../../../lib/api", () => ({
  useApi: () => ({ get, post, put, del: vi.fn() }),
}));

const STORE = "store-1";
const zoneA: Zone = {
  id: "z1",
  name: "Zone A",
  x: 40,
  y: 40,
  width: 200,
  height: 140,
  color: "#6366f1",
  type: "general",
  capacity: 200,
};

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

afterEach(() => vi.clearAllMocks());

describe("useZones", () => {
  it("fetches zones for the store", async () => {
    get.mockResolvedValue([zoneA]);
    const { result } = renderHook(() => useZones(STORE), { wrapper });
    await waitFor(() => expect(result.current.data).toEqual([zoneA]));
    expect(get).toHaveBeenCalledWith(`/api/v1/stores/${STORE}/zones`);
  });

  it("creates a zone via POST", async () => {
    post.mockResolvedValue({ ...zoneA, id: "new" });
    const { result } = renderHook(() => useCreateZone(STORE), { wrapper });
    const body = {
      name: "New Zone",
      x: 0,
      y: 0,
      width: 100,
      height: 80,
      color: "#fff",
      type: "general" as const,
      capacity: 50,
    };
    await result.current.mutateAsync(body);
    expect(post).toHaveBeenCalledWith(`/api/v1/stores/${STORE}/zones`, body);
  });

  it("PUTs the full zone when updating (merges delta)", async () => {
    put.mockResolvedValue({ ...zoneA, x: 80 });
    const { result } = renderHook(() => useUpdateZone(STORE), { wrapper });
    // caller merges the {id,x} delta into the existing zone before calling
    await result.current.mutateAsync({ ...zoneA, x: 80 });
    expect(put).toHaveBeenCalledWith(`/api/v1/stores/${STORE}/zones/z1`, { ...zoneA, x: 80 });
  });
});
