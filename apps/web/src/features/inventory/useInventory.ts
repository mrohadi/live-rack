import { useQuery } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { InventoryRow } from "./types";
import type { ScanRecorded } from "../../lib/ws";

/** Query-key factory — keeps cache keys consistent across hooks. */
export const inventoryKeys = {
  all: ["inventory"] as const,
  list: (storeId: string) => [...inventoryKeys.all, "list", storeId] as const,
};

function inventoryPath(storeId: string): string {
  return `/api/v1/stores/${storeId}/inventory`;
}

/** Fetch current on-hand inventory for a store. */
export function useInventory(storeId: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: inventoryKeys.list(storeId),
    queryFn: () => get<InventoryRow[]>(inventoryPath(storeId)),
  });
}

/** Signed quantity delta a scan applies: pick removes stock, everything else adds. */
export function scanQtyDelta(action: string): number {
  return action === "pick" ? -1 : 1;
}

/** Apply a live scan event to the cached inventory rows. Pure — returns next state. */
export function patchInventory(
  rows: InventoryRow[] | undefined,
  ev: ScanRecorded,
): InventoryRow[] | undefined {
  if (!rows || !ev.valid) return rows;
  return rows.map((r) =>
    r.zone_id === ev.zone_id && r.sku === ev.sku
      ? { ...r, qty: Math.max(0, r.qty + scanQtyDelta(ev.action)), updated_at: ev.ts }
      : r,
  );
}
