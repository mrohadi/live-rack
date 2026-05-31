import { useCallback, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useScanStream } from "../../lib/useScanStream";
import type { ScanRecorded } from "../../lib/ws";
import { useCurrentStore } from "../map/useCurrentStore";
import type { InventoryRow } from "./types";
import { inventoryKeys, patchInventory, useInventory } from "./useInventory";

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

  const [zone, setZone] = useState<string>("all");
  const zones = useMemo(() => Array.from(new Set(rows.map((r) => r.zone_id))), [rows]);
  const visible = zone === "all" ? rows : rows.filter((r) => r.zone_id === zone);

  if (isLoading) {
    return <div style={{ padding: 24, color: "#94a3b8" }}>Loading inventory…</div>;
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <header
        style={{
          padding: "12px 16px",
          borderBottom: "1px solid #1f2937",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <h1 style={{ margin: 0, fontSize: 18 }}>Inventory</h1>
        <select
          data-testid="zone-filter"
          value={zone}
          onChange={(e) => setZone(e.target.value)}
          style={{
            padding: "4px 8px",
            borderRadius: 4,
            border: "1px solid #334155",
            background: "#1e293b",
            color: "#f1f5f9",
            fontSize: 13,
          }}
        >
          <option value="all">All zones</option>
          {zones.map((z) => (
            <option key={z} value={z}>
              {z.slice(0, 8)}
            </option>
          ))}
        </select>
      </header>

      <div style={{ flex: 1, overflow: "auto", padding: 16 }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
          <thead>
            <tr style={{ color: "#94a3b8", textAlign: "left" }}>
              <th style={{ padding: "6px 8px" }}>SKU</th>
              <th style={{ padding: "6px 8px" }}>Name</th>
              <th style={{ padding: "6px 8px" }}>Category</th>
              <th style={{ padding: "6px 8px" }}>Status</th>
              <th style={{ padding: "6px 8px", textAlign: "right" }}>Qty</th>
            </tr>
          </thead>
          <tbody>
            {visible.map((r) => (
              <tr
                key={r.id}
                data-testid="inventory-row"
                style={{ borderTop: "1px solid #1f2937" }}
              >
                <td style={{ padding: "6px 8px", fontFamily: "monospace" }}>{r.sku}</td>
                <td style={{ padding: "6px 8px" }}>{r.name}</td>
                <td style={{ padding: "6px 8px", color: "#94a3b8" }}>{r.category}</td>
                <td style={{ padding: "6px 8px", color: "#94a3b8" }}>{r.status}</td>
                <td
                  data-testid={`qty-${r.sku}`}
                  style={{ padding: "6px 8px", textAlign: "right", fontWeight: 600 }}
                >
                  {r.qty}
                </td>
              </tr>
            ))}
            {visible.length === 0 && (
              <tr>
                <td colSpan={5} style={{ padding: 16, color: "#64748b", textAlign: "center" }}>
                  No stock recorded yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
