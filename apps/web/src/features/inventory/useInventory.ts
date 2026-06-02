import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { InventoryRow, ItemStatus, VelocityBand } from "./types";
import type { ScanRecorded } from "../../lib/ws";

export const ITEM_STATUSES: ItemStatus[] = ["active", "discontinued", "recalled"];
export const VELOCITY_BANDS: VelocityBand[] = ["hot", "warm", "cold", "dead"];

/** Active filter selections. "all" means no constraint on that dimension. */
export interface InventoryFilters {
  zone: string;
  status: string;
  velocity: string;
}

/** Velocity band for a row, defaulting to "cold" until LR-304 populates it. */
export function rowVelocity(r: InventoryRow): VelocityBand {
  return r.velocity ?? "cold";
}

/** Apply zone/status/velocity filters with AND semantics. Pure. */
export function filterInventory(rows: InventoryRow[], f: InventoryFilters): InventoryRow[] {
  return rows.filter(
    (r) =>
      (f.zone === "all" || r.zone_id === f.zone) &&
      (f.status === "all" || r.status === f.status) &&
      (f.velocity === "all" || rowVelocity(r) === f.velocity),
  );
}

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

/** Request body for POST /stores/:storeID/inventory */
export interface AddItemBody {
  zone_id: string;
  sku: string;
  name: string;
  category: string;
  status?: string;
  qty: number;
}

/** Add or adjust an item's qty in a zone. */
export function useAddItem(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: AddItemBody) => post<InventoryRow>(inventoryPath(storeId), body),
    onSuccess: () => qc.invalidateQueries({ queryKey: inventoryKeys.list(storeId) }),
  });
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
