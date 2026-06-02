import { useCallback, useMemo, useState, type ReactNode } from "react";
import { useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useScanStream } from "../../lib/useScanStream";
import type { ScanRecorded } from "../../lib/ws";
import { useCurrentStore } from "../map/useCurrentStore";
import type { InventoryRow } from "./types";
import {
  ITEM_STATUSES,
  VELOCITY_BANDS,
  filterInventory,
  inventoryKeys,
  patchInventory,
  rowVelocity,
  useInventory,
} from "./useInventory";

const VELOCITY_STYLES: Record<string, string> = {
  hot: "bg-destructive/15 text-destructive",
  warm: "bg-warning/15 text-warning",
  cold: "bg-primary/15 text-primary",
  dead: "bg-muted/40 text-muted-foreground",
};

function FilterSelect({
  label,
  value,
  onChange,
  children,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  children: ReactNode;
}) {
  return (
    <label className="flex items-center gap-2 text-xs text-muted-foreground">
      {label}
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="rounded border border-border bg-surface px-2 py-1 text-sm text-foreground"
      >
        {children}
      </select>
    </label>
  );
}

export function InventoryPage() {
  const storeId = useCurrentStore();
  const { data: rows = [], isLoading } = useInventory(storeId);

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

  const zones = useMemo(() => Array.from(new Set(rows.map((r) => r.zone_id))), [rows]);
  const visible = useMemo(
    () => filterInventory(rows, { zone, status, velocity }),
    [rows, zone, status, velocity],
  );

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">Loading inventory…</div>;
  }

  return (
    <div className="flex h-full flex-col">
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
        <h1 className="text-lg font-semibold text-foreground">Inventory</h1>
        <div className="flex flex-wrap items-center gap-3">
          <FilterSelect label="Zone" value={zone} onChange={setZone}>
            <option value="all">All zones</option>
            {zones.map((z) => (
              <option key={z} value={z}>
                {z.slice(0, 8)}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect label="Status" value={status} onChange={setStatus}>
            <option value="all">All</option>
            {ITEM_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect label="Velocity" value={velocity} onChange={setVelocity}>
            <option value="all">All</option>
            {VELOCITY_BANDS.map((v) => (
              <option key={v} value={v}>
                {v}
              </option>
            ))}
          </FilterSelect>
        </div>
      </header>

      <div className="flex-1 overflow-auto p-4">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="text-left text-muted-foreground">
              <th className="px-2 py-1.5 font-medium">SKU</th>
              <th className="px-2 py-1.5 font-medium">Name</th>
              <th className="px-2 py-1.5 font-medium">Category</th>
              <th className="px-2 py-1.5 font-medium">Status</th>
              <th className="px-2 py-1.5 font-medium">Velocity</th>
              <th className="px-2 py-1.5 text-right font-medium">Qty</th>
            </tr>
          </thead>
          <tbody>
            {visible.map((r) => (
              <tr key={r.id} data-testid="inventory-row" className="border-t border-border">
                <td className="px-2 py-1.5 font-mono">{r.sku}</td>
                <td className="px-2 py-1.5 text-foreground">{r.name}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{r.category}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{r.status}</td>
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
              </tr>
            ))}
            {visible.length === 0 && (
              <tr>
                <td colSpan={6} className="p-4 text-center text-muted-foreground">
                  No stock matches the current filters.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
