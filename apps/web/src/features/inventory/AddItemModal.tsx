import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { Select } from "../../components/ui/Select";
import { useZones } from "../map/useZones";
import { useCurrentStore } from "../map/useCurrentStore";
import { ITEM_STATUSES } from "./useInventory";
import { useAddItem } from "./useInventory";

const INPUT =
  "w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20";

const STATUS_OPTIONS = ITEM_STATUSES.map((s) => ({
  value: s,
  label: s.charAt(0).toUpperCase() + s.slice(1),
}));

interface Props {
  /** Pre-select a zone. When set, zone selector is hidden. */
  defaultZoneId?: string;
  onClose: () => void;
}

export function AddItemModal({ defaultZoneId, onClose }: Props) {
  const storeId = useCurrentStore();
  const { data: zones = [] } = useZones(storeId);
  const addItem = useAddItem(storeId);
  const toast = useToast();

  const [sku, setSku] = useState("");
  const [name, setName] = useState("");
  const [category, setCategory] = useState("");
  const [status, setStatus] = useState<string>("active");
  const [zoneId, setZoneId] = useState(defaultZoneId ?? "");
  const [qty, setQty] = useState(1);

  const zoneOptions = zones.map((z) => ({ value: z.id, label: z.name }));

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!sku.trim() || !zoneId || qty <= 0) return;
    addItem.mutate(
      {
        zone_id: zoneId,
        sku: sku.trim(),
        name: name.trim(),
        category: category.trim(),
        status,
        qty,
      },
      {
        onSuccess: () => {
          toast.success(`${sku.trim()} added to zone`);
          onClose();
        },
        onError: () => toast.error("Failed to add item"),
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Add item to zone"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <form
        onSubmit={submit}
        className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-foreground">Add item to zone</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            ✕
          </button>
        </div>

        {/* Zone selector — hidden when defaultZoneId provided */}
        {!defaultZoneId && (
          <Field label="Zone *">
            <Select
              value={zoneId}
              onChange={setZoneId}
              options={zoneOptions}
              placeholder="Select zone…"
            />
          </Field>
        )}

        <Field label="SKU *">
          <input
            required
            value={sku}
            onChange={(e) => setSku(e.target.value)}
            placeholder="e.g. SKU-1234"
            className={INPUT}
          />
        </Field>

        <Field label="Name">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Widget Blue"
            className={INPUT}
          />
        </Field>

        <Field label="Category">
          <input
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            placeholder="e.g. frozen"
            className={INPUT}
          />
        </Field>

        <Field label="Status">
          <Select value={status} onChange={setStatus} options={STATUS_OPTIONS} />
        </Field>

        <Field label="Qty *">
          <input
            required
            type="number"
            min={1}
            value={qty}
            onChange={(e) => setQty(Number(e.target.value))}
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
            disabled={addItem.isPending}
            className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
          >
            {addItem.isPending ? "Adding…" : "Add item"}
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
