import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { useCreateStore, useStores, type StoreItem } from "./useStores";
import { setSelectedStoreId } from "../../lib/storeState";

/** Modal for creating a new store. */
function AddStoreModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [timezone, setTimezone] = useState("UTC");
  const create = useCreateStore();
  const toast = useToast();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    create.mutate(
      { name: trimmed, address: address.trim() || undefined, timezone },
      {
        onSuccess: (s) => {
          toast.success(`Store "${s.name}" created.`);
          onClose();
        },
        onError: () => {
          toast.error("Failed to create store. Try again.");
        },
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      onMouseDown={onClose}
      role="dialog"
      aria-modal="true"
      aria-label="Add store"
    >
      <form
        onSubmit={submit}
        onMouseDown={(e) => e.stopPropagation()}
        className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <h2 className="text-base font-semibold text-foreground">Add store</h2>
        <p className="text-xs text-muted-foreground">
          Each store has its own zones, inventory, and staff assignments.
        </p>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Store name *</span>
          <input
            type="text"
            required
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Main Warehouse"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Address</span>
          <input
            type="text"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder="123 Main St, City"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Timezone</span>
          <select
            value={timezone}
            onChange={(e) => setTimezone(e.target.value)}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
          >
            {TIMEZONES.map((tz) => (
              <option key={tz} value={tz}>
                {tz}
              </option>
            ))}
          </select>
        </label>

        {create.isError && (
          <p role="alert" className="text-xs text-destructive">
            Failed to create store. Please try again.
          </p>
        )}

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground transition hover:bg-muted"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={create.isPending || !name.trim()}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
          >
            {create.isPending ? "Creating…" : "Create store"}
          </button>
        </div>
      </form>
    </div>
  );
}

function StoreCard({
  store,
  active,
  onSelect,
}: {
  store: StoreItem;
  active: boolean;
  onSelect: () => void;
}) {
  return (
    <div
      className={`flex items-start justify-between rounded-lg border p-4 transition ${
        active ? "border-primary bg-primary/5" : "border-border bg-surface hover:border-primary/40"
      }`}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-semibold text-foreground">{store.name}</span>
          {active && (
            <span className="rounded-full bg-primary px-2 py-0.5 text-[10px] font-semibold text-white">
              active
            </span>
          )}
        </div>
        {store.address && (
          <p className="mt-0.5 truncate text-xs text-muted-foreground">{store.address}</p>
        )}
        <p className="mt-0.5 font-mono text-[11px] text-muted-foreground">{store.timezone}</p>
      </div>
      {!active && (
        <button
          type="button"
          onClick={onSelect}
          className="ml-4 flex-none rounded-md border border-border px-2.5 py-1 text-xs font-medium text-foreground transition hover:bg-muted"
        >
          Switch
        </button>
      )}
    </div>
  );
}

export function StoresPage() {
  const { data: stores = [], isLoading } = useStores();
  const [addOpen, setAddOpen] = useState(false);
  const selectedId = localStorage.getItem("lr:store_id");
  const toast = useToast();

  const handleSelect = (id: string, name: string) => {
    setSelectedStoreId(id);
    toast.success(`Switched to "${name}".`);
    setTimeout(() => window.location.reload(), 800);
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-foreground">Stores</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Manage warehouse locations for your organisation.
          </p>
        </div>
        <button
          type="button"
          onClick={() => setAddOpen(true)}
          className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-white transition hover:opacity-90"
        >
          + Add store
        </button>
      </div>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : stores.length === 0 ? (
        <div className="rounded-lg border border-dashed border-border p-10 text-center">
          <p className="text-sm text-muted-foreground">No stores yet.</p>
          <button
            type="button"
            onClick={() => setAddOpen(true)}
            className="mt-3 text-sm font-medium text-primary hover:underline"
          >
            Create your first store →
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {stores.map((s) => (
            <StoreCard
              key={s.id}
              store={s}
              active={s.id === (selectedId ?? stores[0]?.id)}
              onSelect={() => handleSelect(s.id, s.name)}
            />
          ))}
        </div>
      )}

      {addOpen && <AddStoreModal onClose={() => setAddOpen(false)} />}
    </div>
  );
}

const TIMEZONES = [
  "UTC",
  "Asia/Jakarta",
  "Asia/Singapore",
  "Asia/Tokyo",
  "Asia/Shanghai",
  "Asia/Kolkata",
  "Asia/Dubai",
  "Europe/London",
  "Europe/Berlin",
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "Australia/Sydney",
];
