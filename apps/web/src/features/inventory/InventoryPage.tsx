import { useCallback, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Select } from "../../components/ui/Select";
import { useScanStream } from "../../lib/useScanStream";
import type { ScanRecorded } from "../../lib/ws";
import { useCurrentStore } from "../map/useCurrentStore";
import { useZones } from "../map/useZones";
import type { InventoryRow } from "./types";
import {
  ITEM_STATUSES,
  VELOCITY_BANDS,
  filterInventory,
  formatCents,
  inventoryKeys,
  patchInventory,
  rowStockStatus,
  rowVelocity,
  useInventory,
} from "./useInventory";
import { AddItemModal } from "./AddItemModal";
import { TransferStockModal } from "./TransferStockModal";
import { ItemDetailDrawer } from "./ItemDetailDrawer";

const VELOCITY_STYLES: Record<string, string> = {
  hot: "bg-destructive/15 text-destructive",
  warm: "bg-warning/15 text-warning",
  cold: "bg-primary/15 text-primary",
  dead: "bg-muted/40 text-muted-foreground",
};

const STOCK_STYLES: Record<string, string> = {
  in_stock: "bg-primary/15 text-primary",
  low: "bg-warning/15 text-warning",
  out: "bg-destructive/15 text-destructive",
};

const STOCK_LABELS: Record<string, string> = {
  in_stock: "In stock",
  low: "Low",
  out: "Out",
};

const STOCK_TABS: { value: string; label: string }[] = [
  { value: "all", label: "All" },
  { value: "in_stock", label: "In stock" },
  { value: "low", label: "Low" },
  { value: "out", label: "Out" },
];

const STATUS_OPTIONS = [
  { value: "all", label: "All statuses" },
  ...ITEM_STATUSES.map((s) => ({
    value: s,
    label: s.charAt(0).toUpperCase() + s.slice(1),
  })),
];

const VELOCITY_OPTIONS = [
  { value: "all", label: "All velocities" },
  ...VELOCITY_BANDS.map((v) => ({
    value: v,
    label: v.charAt(0).toUpperCase() + v.slice(1),
    badge: v,
    badgeClass: VELOCITY_STYLES[v],
  })),
];

export function InventoryPage() {
  const storeId = useCurrentStore();
  const { data: rows = [], isLoading } = useInventory(storeId);
  const { data: zones = [] } = useZones(storeId);

  const qc = useQueryClient();
  const onScan = useCallback(
    (ev: ScanRecorded) => {
      qc.setQueryData<InventoryRow[]>(inventoryKeys.list(storeId), (prev) =>
        patchInventory(prev, ev),
      );
    },
    [qc, storeId],
  );
  useScanStream(onScan);

  const [searchParams] = useSearchParams();
  const [zone, setZone] = useState(searchParams.get("zone") ?? "all");
  const [status, setStatus] = useState("all");
  const [velocity, setVelocity] = useState("all");
  const [stock, setStock] = useState("all");
  const [showAdd, setShowAdd] = useState(false);
  const [transferRow, setTransferRow] = useState<InventoryRow | null>(null);
  const [detailSku, setDetailSku] = useState<string | null>(null);

  const zoneOptions = useMemo(
    () => [
      { value: "all", label: "All zones" },
      ...zones.map((z) => ({ value: z.id, label: z.name })),
    ],
    [zones],
  );

  const visible = useMemo(
    () => filterInventory(rows, { zone, status, velocity, stock }),
    [rows, zone, status, velocity, stock],
  );

  const lowCount = useMemo(
    () => rows.filter((r) => rowStockStatus(r) !== "in_stock").length,
    [rows],
  );

  const visibleValue = useMemo(
    () => visible.reduce((sum, r) => sum + (r.value_cents ?? 0), 0),
    [visible],
  );

  const zoneNameById = useMemo(() => Object.fromEntries(zones.map((z) => [z.id, z.name])), [zones]);

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">Loading inventory…</div>;
  }

  return (
    <div className="flex h-full flex-col">
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div className="flex items-baseline gap-3">
          <h1 className="text-lg font-semibold text-foreground">Inventory</h1>
          <span className="text-xs text-muted-foreground">{formatCents(visibleValue)} on hand</span>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Zone</span>
            <Select value={zone} onChange={setZone} options={zoneOptions} className="w-44" />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Status</span>
            <Select value={status} onChange={setStatus} options={STATUS_OPTIONS} className="w-40" />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Velocity</span>
            <Select
              value={velocity}
              onChange={setVelocity}
              options={VELOCITY_OPTIONS}
              className="w-40"
            />
          </div>
          <button
            type="button"
            onClick={() => setShowAdd(true)}
            className="rounded-lg bg-primary px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
          >
            + Add item
          </button>
        </div>
      </header>

      <div className="flex items-center gap-2 border-b border-border px-4 py-2">
        {STOCK_TABS.map((t) => (
          <button
            key={t.value}
            type="button"
            onClick={() => setStock(t.value)}
            className={`rounded-full px-3 py-1 text-xs font-medium transition ${
              stock === t.value
                ? "bg-primary text-white"
                : "bg-muted/40 text-muted-foreground hover:bg-muted"
            }`}
          >
            {t.label}
          </button>
        ))}
        {lowCount > 0 && (
          <span className="ml-auto text-xs text-warning">{lowCount} need attention</span>
        )}
      </div>

      <div className="flex-1 overflow-auto p-4">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="text-left text-muted-foreground">
              <th className="px-2 py-1.5 font-medium">SKU</th>
              <th className="px-2 py-1.5 font-medium">Name</th>
              <th className="px-2 py-1.5 font-medium">Category</th>
              <th className="px-2 py-1.5 font-medium">Zone</th>
              <th className="px-2 py-1.5 font-medium">Status</th>
              <th className="px-2 py-1.5 font-medium">Stock</th>
              <th className="px-2 py-1.5 font-medium">Velocity</th>
              <th className="px-2 py-1.5 text-right font-medium">Qty</th>
              <th className="px-2 py-1.5 text-right font-medium">Value</th>
              <th className="px-2 py-1.5" />
            </tr>
          </thead>
          <tbody>
            {visible.map((r) => (
              <tr
                key={r.id}
                data-testid="inventory-row"
                onClick={() => setDetailSku(r.sku)}
                className="cursor-pointer border-t border-border hover:bg-muted/40"
              >
                <td className="px-2 py-1.5 font-mono text-xs">{r.sku}</td>
                <td className="px-2 py-1.5 text-foreground">{r.name}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{r.category}</td>
                <td className="px-2 py-1.5 text-muted-foreground">
                  {zoneNameById[r.zone_id] ?? r.zone_id.slice(0, 8)}
                </td>
                <td className="px-2 py-1.5 text-muted-foreground">{r.status}</td>
                <td className="px-2 py-1.5">
                  <span
                    className={`rounded px-1.5 py-0.5 text-xs ${STOCK_STYLES[rowStockStatus(r)]}`}
                  >
                    {STOCK_LABELS[rowStockStatus(r)]}
                  </span>
                </td>
                <td className="px-2 py-1.5">
                  <span
                    className={`rounded px-1.5 py-0.5 text-xs ${VELOCITY_STYLES[rowVelocity(r)]}`}
                  >
                    {rowVelocity(r)}
                  </span>
                </td>
                <td
                  data-testid={`qty-${r.sku}`}
                  className="px-2 py-1.5 text-right font-semibold text-foreground"
                >
                  {r.qty}
                </td>
                <td className="px-2 py-1.5 text-right text-muted-foreground">
                  {formatCents(r.value_cents)}
                </td>
                <td className="px-2 py-1.5 text-right">
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      setTransferRow(r);
                    }}
                    disabled={r.qty <= 0}
                    className="rounded-md border border-border px-2 py-1 text-xs text-muted-foreground transition hover:bg-muted hover:text-foreground disabled:opacity-40"
                  >
                    ⇄ Transfer
                  </button>
                </td>
              </tr>
            ))}
            {visible.length === 0 && (
              <tr>
                <td colSpan={10} className="p-4 text-center text-muted-foreground">
                  No stock matches the current filters.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {showAdd && <AddItemModal onClose={() => setShowAdd(false)} />}
      {transferRow && <TransferStockModal row={transferRow} onClose={() => setTransferRow(null)} />}
      {detailSku && <ItemDetailDrawer sku={detailSku} onClose={() => setDetailSku(null)} />}
    </div>
  );
}
