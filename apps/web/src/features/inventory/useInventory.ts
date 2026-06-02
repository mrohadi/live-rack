import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { InventoryRow, ItemDetail, ItemStatus, StockStatus, VelocityBand } from "./types";
import type { ScanRecorded } from "../../lib/ws";

export const ITEM_STATUSES: ItemStatus[] = ["active", "discontinued", "recalled"];
export const VELOCITY_BANDS: VelocityBand[] = ["hot", "warm", "cold", "dead"];
export const STOCK_STATUSES: StockStatus[] = ["in_stock", "low", "out"];

/** Format integer cents as a USD string. Pure. */
export function formatCents(cents: number | undefined): string {
  return `$${((cents ?? 0) / 100).toFixed(2)}`;
}

/** Derive on-hand stock band from qty vs reorder point. Pure. Mirrors the API. */
export function rowStockStatus(r: InventoryRow): StockStatus {
  if (r.stock_status) return r.stock_status;
  const rp = r.reorder_point ?? 0;
  if (r.qty <= 0) return "out";
  if (rp > 0 && r.qty <= rp) return "low";
  return "in_stock";
}

/** Active filter selections. "all" means no constraint on that dimension. */
export interface InventoryFilters {
  zone: string;
  status: string;
  velocity: string;
  /** Stock band: "all" | in_stock | low | out (LR-305). */
  stock?: string;
}

/** Velocity band for a row, defaulting to "cold" until LR-304 populates it. */
export function rowVelocity(r: InventoryRow): VelocityBand {
  return r.velocity ?? "cold";
}

/** Apply zone/status/velocity/stock filters with AND semantics. Pure. */
export function filterInventory(rows: InventoryRow[], f: InventoryFilters): InventoryRow[] {
  const stock = f.stock ?? "all";
  return rows.filter(
    (r) =>
      (f.zone === "all" || r.zone_id === f.zone) &&
      (f.status === "all" || r.status === f.status) &&
      (f.velocity === "all" || rowVelocity(r) === f.velocity) &&
      (stock === "all" || rowStockStatus(r) === stock),
  );
}

/** Query-key factory — keeps cache keys consistent across hooks. */
export const inventoryKeys = {
  all: ["inventory"] as const,
  list: (storeId: string) => [...inventoryKeys.all, "list", storeId] as const,
  detail: (storeId: string, sku: string) => [...inventoryKeys.all, "detail", storeId, sku] as const,
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

/** Fetch full item detail (per-zone on-hand + scan timeline) for the drawer. */
export function useItemDetail(storeId: string, sku: string | null) {
  const { get } = useApi();
  return useQuery({
    queryKey: inventoryKeys.detail(storeId, sku ?? ""),
    queryFn: () => get<ItemDetail>(`${inventoryPath(storeId)}/${encodeURIComponent(sku ?? "")}`),
    enabled: !!sku,
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
  /** Restock trigger threshold; 0 disables (LR-305). */
  reorder_point?: number;
  /** Unit price in integer cents (LR-309). */
  price_cents?: number;
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

/** Request body for POST /stores/:storeID/inventory/transfer (LR-306). */
export interface TransferStockBody {
  sku: string;
  from_zone_id: string;
  to_zone_id: string;
  qty: number;
}

/** Move stock of a SKU from one zone to another. */
export function useTransferStock(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: TransferStockBody) =>
      post<TransferStockBody>(`${inventoryPath(storeId)}/transfer`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: inventoryKeys.list(storeId) }),
  });
}

/** Editable master-catalog fields (LR-310). */
export interface EditItemBody {
  name: string;
  category: string;
  status: string;
  reorder_point: number;
  price_cents: number;
}

function invalidateItem(qc: ReturnType<typeof useQueryClient>, storeId: string, sku: string) {
  void qc.invalidateQueries({ queryKey: inventoryKeys.list(storeId) });
  void qc.invalidateQueries({ queryKey: inventoryKeys.detail(storeId, sku) });
}

/** Edit catalog fields for a SKU. */
export function useEditItem(storeId: string, sku: string) {
  const { patch } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: EditItemBody) =>
      patch<EditItemBody>(`${inventoryPath(storeId)}/${encodeURIComponent(sku)}`, body),
    onSuccess: () => invalidateItem(qc, storeId, sku),
  });
}

/** Absolute on-hand correction for one zone (shrinkage, damage, count). */
export function useAdjustQty(storeId: string, sku: string) {
  const { patch } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { zone_id: string; qty: number }) =>
      patch(`${inventoryPath(storeId)}/${encodeURIComponent(sku)}/qty`, body),
    onSuccess: () => invalidateItem(qc, storeId, sku),
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
