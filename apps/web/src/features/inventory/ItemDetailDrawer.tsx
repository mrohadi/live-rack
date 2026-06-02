import { useCurrentStore } from "../map/useCurrentStore";
import type { StockStatus } from "./types";
import { useItemDetail } from "./useInventory";

const STOCK_STYLES: Record<StockStatus, string> = {
  in_stock: "bg-primary/15 text-primary",
  low: "bg-warning/15 text-warning",
  out: "bg-destructive/15 text-destructive",
};

const STOCK_LABELS: Record<StockStatus, string> = {
  in_stock: "In stock",
  low: "Low",
  out: "Out",
};

interface Props {
  sku: string;
  onClose: () => void;
}

export function ItemDetailDrawer({ sku, onClose }: Props) {
  const storeId = useCurrentStore();
  const { data, isLoading } = useItemDetail(storeId, sku);

  return (
    <div
      className="fixed inset-0 z-50 flex justify-end bg-black/40"
      role="dialog"
      aria-modal="true"
      aria-label="Item detail"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <aside className="flex h-full w-full max-w-md flex-col border-l border-border bg-surface shadow-xl">
        <header className="flex items-start justify-between border-b border-border px-5 py-4">
          <div className="min-w-0">
            <p className="font-mono text-xs text-muted-foreground">{sku}</p>
            <h2 className="truncate text-base font-semibold text-foreground">
              {data?.name || (isLoading ? "Loading…" : sku)}
            </h2>
            {data?.category && <p className="text-xs text-muted-foreground">{data.category}</p>}
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            ✕
          </button>
        </header>

        {isLoading || !data ? (
          <div className="p-5 text-sm text-muted-foreground">Loading item…</div>
        ) : (
          <div className="flex-1 space-y-6 overflow-auto p-5">
            {/* Summary */}
            <div className="grid grid-cols-3 gap-3">
              <Stat label="Total qty" value={String(data.total_qty)} />
              <Stat
                label="Reorder pt"
                value={data.reorder_point > 0 ? String(data.reorder_point) : "—"}
              />
              <div>
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Stock</p>
                <span
                  className={`mt-1 inline-block rounded px-1.5 py-0.5 text-xs ${STOCK_STYLES[data.stock_status]}`}
                >
                  {STOCK_LABELS[data.stock_status]}
                </span>
              </div>
            </div>

            {/* Per-zone locations */}
            <section>
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Locations
              </h3>
              <div className="overflow-hidden rounded-lg border border-border">
                <table className="w-full text-sm">
                  <thead className="bg-muted/40 text-left text-muted-foreground">
                    <tr>
                      <th className="px-3 py-1.5 font-medium">Zone</th>
                      <th className="px-3 py-1.5 font-medium">Stock</th>
                      <th className="px-3 py-1.5 text-right font-medium">Qty</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.locations.map((l) => (
                      <tr key={l.zone_id} className="border-t border-border">
                        <td className="px-3 py-1.5 text-foreground">
                          {l.zone_name || l.zone_id.slice(0, 8)}
                        </td>
                        <td className="px-3 py-1.5">
                          <span
                            className={`rounded px-1.5 py-0.5 text-xs ${STOCK_STYLES[l.stock_status]}`}
                          >
                            {STOCK_LABELS[l.stock_status]}
                          </span>
                        </td>
                        <td className="px-3 py-1.5 text-right font-semibold text-foreground">
                          {l.qty}
                        </td>
                      </tr>
                    ))}
                    {data.locations.length === 0 && (
                      <tr>
                        <td colSpan={3} className="px-3 py-3 text-center text-muted-foreground">
                          No stock on hand.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </section>

            {/* Scan timeline */}
            <section>
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Recent scans
              </h3>
              <ul className="space-y-2">
                {data.recent_scans.map((s, i) => (
                  <li key={i} className="flex items-center gap-2 text-sm">
                    <span
                      className={`h-2 w-2 shrink-0 rounded-full ${s.valid ? "bg-primary" : "bg-destructive"}`}
                    />
                    <span className="font-medium text-foreground">{s.action}</span>
                    <span className="text-muted-foreground">{s.scanner_id}</span>
                    <span className="ml-auto text-xs text-muted-foreground">
                      {new Date(s.ts).toLocaleString()}
                    </span>
                  </li>
                ))}
                {data.recent_scans.length === 0 && (
                  <li className="text-sm text-muted-foreground">No scans yet.</li>
                )}
              </ul>
            </section>
          </div>
        )}
      </aside>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs uppercase tracking-wide text-muted-foreground">{label}</p>
      <p className="mt-1 text-lg font-semibold text-foreground">{value}</p>
    </div>
  );
}
