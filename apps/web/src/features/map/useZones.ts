import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { Zone } from "./types";

/** Query-key factory — keeps cache keys consistent across hooks. */
export const zoneKeys = {
  all: ["zones"] as const,
  list: (storeId: string) => [...zoneKeys.all, "list", storeId] as const,
};

function zonesPath(storeId: string): string {
  return `/api/v1/stores/${storeId}/zones`;
}

/** Fetch all zones for a store. */
export function useZones(storeId: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: zoneKeys.list(storeId),
    queryFn: () => get<Zone[]>(zonesPath(storeId)),
  });
}

/** Create a new zone in a store. */
export function useCreateZone(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: Omit<Zone, "id">) => post<Zone>(zonesPath(storeId), body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: zoneKeys.list(storeId) });
    },
  });
}

/**
 * Update a zone. The backend PUT validates and requires a FULL zone body,
 * so callers must pass the complete merged zone — not a partial delta.
 */
export function useUpdateZone(storeId: string) {
  const { put } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (zone: Zone) => put<Zone>(`${zonesPath(storeId)}/${zone.id}`, zone),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: zoneKeys.list(storeId) });
    },
  });
}
