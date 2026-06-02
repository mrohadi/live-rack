import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { Select } from "../../components/ui/Select";
import { useCreateZone } from "./useZones";
import { useCurrentStore } from "./useCurrentStore";
import { findOpenSlot, randomZoneColor } from "./zoneMath";
import type { Zone } from "./types";

const ZONE_TYPE_OPTIONS = [
  { value: "general", label: "General" },
  { value: "frozen", label: "Frozen" },
  { value: "returns", label: "Returns" },
  { value: "staging", label: "Staging" },
  { value: "display", label: "Display" },
  { value: "checkout", label: "Checkout" },
];

interface Props {
  zones: Zone[];
  onClose: () => void;
}

export function ZoneCreateModal({ zones, onClose }: Props) {
  const storeId = useCurrentStore();
  const createZone = useCreateZone(storeId);
  const toast = useToast();

  const [name, setName] = useState("");
  const [type, setType] = useState<Zone["type"]>("general");
  const [capacity, setCapacity] = useState("100");

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    const cap = parseInt(capacity, 10);
    const width = 20;
    const height = 14;
    const { x, y } = findOpenSlot(zones, width, height);
    createZone.mutate(
      {
        name: name.trim(),
        type,
        x,
        y,
        width,
        height,
        color: randomZoneColor(),
        capacity: isNaN(cap) || cap < 0 ? 100 : cap,
      },
      {
        onSuccess: () => {
          toast.success(`Zone "${name.trim()}" created`);
          onClose();
        },
        onError: () => toast.error("Failed to create zone"),
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Create zone"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <form
        onSubmit={submit}
        className="w-full max-w-sm space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-foreground">New zone</h2>
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
            placeholder="e.g. Apparel"
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
          />
        </Field>

        <Field label="Type">
          <Select
            value={type}
            onChange={(v) => setType(v as Zone["type"])}
            options={ZONE_TYPE_OPTIONS}
          />
        </Field>

        <Field label="Capacity">
          <input
            type="number"
            min={0}
            value={capacity}
            onChange={(e) => setCapacity(e.target.value)}
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
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
            disabled={createZone.isPending}
            className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
          >
            {createZone.isPending ? "Creating…" : "Create zone"}
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
