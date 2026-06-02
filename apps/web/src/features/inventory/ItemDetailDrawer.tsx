import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { Select } from "../../components/ui/Select";
import { useCurrentStore } from "../map/useCurrentStore";
import type { ItemDetail, StockStatus } from "./types";
import { ITEM_STATUSES, useAdjustQty, useEditItem, useItemDetail } from "./useInventory";

const INPUT =
  "w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20";

const STATUS_OPTIONS = ITEM_STATUSES.map((s) => ({
  value: s,
  label: s.charAt(0).toUpperCase() + s.slice(1),
}));

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
  const [editing, setEditing] = useState(false);

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
          <div className="flex items-center gap-2">
            {data && !editing && (
              <button
                type="button"
                onClick={() => setEditing(true)}
                className="rounded-md border border-border px-2 py-1 text-xs text-foreground transition hover:bg-muted"
              >
                Edit
              </button>
            )}
            <button
              type="button"
              onClick={onClose}
              aria-label="Close"
              className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              ✕
            </button>
          </div>
        </header>

        {isLoading || !data ? (
          <div className="p-5 text-sm text-muted-foreground">Loading item…</div>
        ) : editing ? (
          <EditForm storeId={storeId} sku={sku} detail={data} onDone={() => setEditing(false)} />
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
                        <AdjustQtyCell storeId={storeId} sku={sku} zoneId={l.zone_id} qty={l.qty} />
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

function EditForm({
  storeId,
  sku,
  detail,
  onDone,
}: {
  storeId: string;
  sku: string;
  detail: ItemDetail;
  onDone: () => void;
}) {
  const edit = useEditItem(storeId, sku);
  const toast = useToast();

  const [name, setName] = useState(detail.name);
  const [category, setCategory] = useState(detail.category);
  const [status, setStatus] = useState(detail.status);
  const [reorderPoint, setReorderPoint] = useState(detail.reorder_point);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    edit.mutate(
      { name: name.trim(), category: category.trim(), status, reorder_point: reorderPoint },
      {
        onSuccess: () => {
          toast.success("Item updated");
          onDone();
        },
        onError: () => toast.error("Failed to update item"),
      },
    );
  };

  return (
    <form onSubmit={submit} className="flex-1 space-y-4 overflow-auto p-5">
      <Field label="Name">
        <input value={name} onChange={(e) => setName(e.target.value)} className={INPUT} />
      </Field>
      <Field label="Category">
        <input value={category} onChange={(e) => setCategory(e.target.value)} className={INPUT} />
      </Field>
      <Field label="Status">
        <Select value={status} onChange={setStatus} options={STATUS_OPTIONS} />
      </Field>
      <Field label="Reorder point (0 = off)">
        <input
          type="number"
          min={0}
          value={reorderPoint}
          onChange={(e) => setReorderPoint(Math.max(0, Number(e.target.value)))}
          className={INPUT}
        />
      </Field>
      <div className="flex justify-end gap-2 pt-1">
        <button
          type="button"
          onClick={onDone}
          className="rounded-lg border border-border px-3 py-1.5 text-sm text-foreground transition hover:bg-muted"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={edit.isPending}
          className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
        >
          {edit.isPending ? "Saving…" : "Save"}
        </button>
      </div>
    </form>
  );
}

function AdjustQtyCell({
  storeId,
  sku,
  zoneId,
  qty,
}: {
  storeId: string;
  sku: string;
  zoneId: string;
  qty: number;
}) {
  const adjust = useAdjustQty(storeId, sku);
  const toast = useToast();
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(qty);

  const save = () => {
    adjust.mutate(
      { zone_id: zoneId, qty: Math.max(0, value) },
      {
        onSuccess: () => {
          toast.success("Quantity corrected");
          setEditing(false);
        },
        onError: () => toast.error("Failed to set quantity"),
      },
    );
  };

  if (!editing) {
    return (
      <td className="px-3 py-1.5 text-right">
        <button
          type="button"
          onClick={() => {
            setValue(qty);
            setEditing(true);
          }}
          className="font-semibold text-foreground underline-offset-2 hover:underline"
          title="Correct on-hand qty"
        >
          {qty}
        </button>
      </td>
    );
  }

  return (
    <td className="px-3 py-1.5">
      <div className="flex items-center justify-end gap-1">
        <input
          type="number"
          min={0}
          value={value}
          autoFocus
          onChange={(e) => setValue(Math.max(0, Number(e.target.value)))}
          className="w-16 rounded border border-border bg-background px-1.5 py-0.5 text-right text-sm text-foreground outline-none focus:border-primary"
        />
        <button
          type="button"
          onClick={save}
          disabled={adjust.isPending}
          className="rounded bg-primary px-1.5 py-0.5 text-xs text-white disabled:opacity-50"
        >
          ✓
        </button>
        <button
          type="button"
          onClick={() => setEditing(false)}
          className="rounded border border-border px-1.5 py-0.5 text-xs text-muted-foreground"
        >
          ✕
        </button>
      </div>
    </td>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block text-sm">
      <span className="mb-1.5 block text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </span>
      {children}
    </label>
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
