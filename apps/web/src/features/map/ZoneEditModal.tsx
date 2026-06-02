import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { useUpdateZone } from "./useZones";
import { useCurrentStore } from "./useCurrentStore";
import type { Zone } from "./types";

interface Props {
  zone: Zone;
  onClose: () => void;
}

export function ZoneEditModal({ zone, onClose }: Props) {
  const storeId = useCurrentStore();
  const updateZone = useUpdateZone(storeId);
  const toast = useToast();

  const [name, setName] = useState(zone.name);
  const [capacity, setCapacity] = useState(String(zone.capacity ?? 100));
  const [cats, setCats] = useState(zone.constraints?.allowedCategories?.join(", ") ?? "");
  const [maxSKUs, setMaxSKUs] = useState(
    zone.constraints?.maxSKUs != null ? String(zone.constraints.maxSKUs) : "",
  );
  const [climate, setClimate] = useState(zone.constraints?.climate ?? "");

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    const cap = parseInt(capacity, 10);
    if (isNaN(cap) || cap < 0) return;

    const parsedCats = cats
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    const parsedMaxSKUs = maxSKUs ? parseInt(maxSKUs, 10) : undefined;
    const parsedClimate = climate.trim() || undefined;

    const constraints =
      parsedCats.length || parsedMaxSKUs != null || parsedClimate
        ? {
            ...(parsedCats.length ? { allowedCategories: parsedCats } : {}),
            ...(!isNaN(parsedMaxSKUs as number) && parsedMaxSKUs != null
              ? { maxSKUs: parsedMaxSKUs }
              : {}),
            ...(parsedClimate ? { climate: parsedClimate } : {}),
          }
        : undefined;

    updateZone.mutate(
      { ...zone, name: name.trim(), capacity: cap, constraints },
      {
        onSuccess: () => {
          toast.success("Zone updated");
          onClose();
        },
        onError: () => toast.error("Failed to update zone"),
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Edit zone"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <form
        onSubmit={submit}
        className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-foreground">Edit zone</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            ✕
          </button>
        </div>

        <Field label="Zone name *">
          <input
            autoFocus
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <Field label="Capacity *">
          <input
            required
            type="number"
            min={0}
            value={capacity}
            onChange={(e) => setCapacity(e.target.value)}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <div className="border-t border-border pt-3">
          <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Constraints
          </p>
          <div className="space-y-3">
            <Field label="Allowed categories (comma-separated)">
              <input
                value={cats}
                onChange={(e) => setCats(e.target.value)}
                placeholder="e.g. frozen, apparel"
                className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
            </Field>
            <Field label="Max SKUs (leave blank for unlimited)">
              <input
                type="number"
                min={0}
                value={maxSKUs}
                onChange={(e) => setMaxSKUs(e.target.value)}
                placeholder="unlimited"
                className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
            </Field>
            <Field label="Climate control">
              <input
                value={climate}
                onChange={(e) => setClimate(e.target.value)}
                placeholder="e.g. refrigerated, ambient"
                className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
            </Field>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground hover:bg-muted"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={updateZone.isPending}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            {updateZone.isPending ? "Saving…" : "Save changes"}
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
