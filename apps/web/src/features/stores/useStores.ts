import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import { useEffect } from "react";
import { getSelectedStoreId, setSelectedStoreId } from "../../lib/storeState";

export interface StoreItem {
  id: string;
  name: string;
  address?: string;
  timezone: string;
}

const STORES_KEY = ["stores"] as const;

export function useStores() {
  const { get } = useApi();
  return useQuery({
    queryKey: STORES_KEY,
    queryFn: () => get<StoreItem[]>("/api/v1/stores"),
    staleTime: 60_000,
  });
}

/** Returns the selected store, falling back to the first available one. */
export function useCurrentStoreData(): StoreItem | undefined {
  const { data: stores = [] } = useStores();
  const saved = getSelectedStoreId();
  // Auto-select the first store when nothing is persisted yet.
  useEffect(() => {
    if (stores.length > 0 && !saved) {
      setSelectedStoreId(stores[0].id);
    }
  }, [stores, saved]);
  const selected = saved ? stores.find((s) => s.id === saved) : undefined;
  return selected ?? stores[0];
}

export function useCreateStore() {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string; address?: string; timezone?: string }) =>
      post<StoreItem>("/api/v1/stores", body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: STORES_KEY }),
  });
}
