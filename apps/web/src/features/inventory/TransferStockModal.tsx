import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { Select } from "../../components/ui/Select";
import { useZones } from "../map/useZones";
import { useCurrentStore } from "../map/useCurrentStore";
import type { InventoryRow } from "./types";
import { useTransferStock } from "./useInventory";

const INPUT =
  "w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20";

interface Props {
  /** Source row: fixes SKU + from-zone, caps qty at on-hand. */
  row: InventoryRow;
  onClose: () => void;
}

export function TransferStockModal({ row, onClose }: Props) {
  const storeId = useCurrentStore();
  const { data: zones = [] } = useZones(storeId);
  const transfer = useTransferStock(storeId);
  const toast = useToast();

  const [toZone, setToZone] = useState("");
  const [qty, setQty] = useState(1);

  const fromZoneName = zones.find((z) => z.id === row.zone_id)?.name ?? row.zone_id.slice(0, 8);
  const toZoneOptions = zones
    .filter((z) => z.id !== row.zone_id)
    .map((z) => ({ value: z.id, label: z.name }));

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!toZone || qty <= 0 || qty > row.qty) return;
    transfer.mutate(
      { sku: row.sku, from_zone_id: row.zone_id, to_zone_id: toZone, qty },
      {
        onSuccess: () => {
          toast.success(`Moved ${qty} × ${row.sku}`);
          onClose();
        },
        onError: () => toast.error("Transfer failed — check stock"),
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Transfer stock"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <form
        onSubmit={submit}
        className="w-full max-w-sm space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-foreground">Transfer stock</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            ✕
          </button>
        </div>

        <div className="flex items-center gap-2 rounded-lg bg-muted px-3 py-2 text-sm">
          <span className="font-mono text-xs text-foreground">{row.sku}</span>
          <span className="text-muted-foreground">{row.name}</span>
          <span className="ml-auto text-xs text-muted-foreground">{row.qty} on hand</span>
        </div>

        <Field label="From zone">
          <div className={`${INPUT} cursor-not-allowed bg-muted/40 text-muted-foreground`}>
            {fromZoneName}
          </div>
        </Field>

        <Field label="To zone *">
          <Select
            value={toZone}
            onChange={setToZone}
            options={toZoneOptions}
            placeholder="Select destination…"
          />
        </Field>

        <Field label="Qty *">
          <input
            required
            type="number"
            min={1}
            max={row.qty}
            value={qty}
            onChange={(e) => setQty(Math.max(1, Math.min(row.qty, Number(e.target.value))))}
            className={INPUT}
          />
        </Field>

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-border px-3 py-1.5 text-sm text-foreground transition hover:bg-muted"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={transfer.isPending || !toZone || qty > row.qty}
            className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
          >
            {transfer.isPending ? "Moving…" : "Move stock"}
          </button>
        </div>
      </form>
    </div>
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
