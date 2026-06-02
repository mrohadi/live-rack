import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { useZones } from "../map/useZones";
import { useCurrentStore } from "../map/useCurrentStore";
import { ITEM_STATUSES } from "./useInventory";
import { useAddItem } from "./useInventory";

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
    >
      <form
        onSubmit={submit}
        className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <h2 className="text-base font-semibold text-foreground">Add item to zone</h2>

        {/* Zone selector — hidden when defaultZoneId provided */}
        {!defaultZoneId && (
          <Field label="Zone">
            <select
              required
              value={zoneId}
              onChange={(e) => setZoneId(e.target.value)}
              className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
            >
              <option value="">Select zone…</option>
              {zones.map((z) => (
                <option key={z.id} value={z.id}>
                  {z.name}
                </option>
              ))}
            </select>
          </Field>
        )}

        <Field label="SKU *">
          <input
            required
            value={sku}
            onChange={(e) => setSku(e.target.value)}
            placeholder="e.g. SKU-1234"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <Field label="Name">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Widget Blue"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <Field label="Category">
          <input
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            placeholder="e.g. frozen"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <Field label="Status">
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          >
            {ITEM_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </Field>

        <Field label="Qty *">
          <input
            required
            type="number"
            min={1}
            value={qty}
            onChange={(e) => setQty(Number(e.target.value))}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={addItem.isPending}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
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
      <span className="mb-1 block text-muted-foreground">{label}</span>
      {children}
    </label>
  );
}
