import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import type { NewPickLine } from "./types";
import { useCreatePickList } from "./usePicking";

interface Props {
  storeId: string;
  onCreated: (id: string) => void;
  onCancel: () => void;
}

export function NewPickListForm({ storeId, onCreated, onCancel }: Props) {
  const toast = useToast();
  const create = useCreatePickList(storeId);
  const [reference, setReference] = useState("");
  const [lines, setLines] = useState<NewPickLine[]>([{ sku: "", qty: 1 }]);

  function update(i: number, patch: Partial<NewPickLine>) {
    setLines((cur) => cur.map((l, idx) => (idx === i ? { ...l, ...patch } : l)));
  }

  function submit() {
    const clean = lines
      .map((l) => ({ sku: l.sku.trim(), qty: Number(l.qty) }))
      .filter((l) => l.sku !== "" && l.qty > 0);
    if (clean.length === 0) {
      toast.error("Add at least one SKU with a quantity");
      return;
    }
    create.mutate(
      { reference: reference.trim(), lines: clean },
      {
        onSuccess: (board) => {
          toast.success("Pick list created");
          onCreated(board.id);
        },
        onError: () => toast.error("Failed to create pick list"),
      },
    );
  }

  return (
    <div className="space-y-3 rounded-lg border border-border bg-card p-3">
      <input
        value={reference}
        onChange={(e) => setReference(e.target.value)}
        placeholder="Reference (e.g. SO-1042)"
        aria-label="Pick list reference"
        className="w-full rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
      />
      <div className="space-y-2">
        {lines.map((l, i) => (
          <div key={i} className="flex items-center gap-2">
            <input
              value={l.sku}
              onChange={(e) => update(i, { sku: e.target.value })}
              placeholder="SKU"
              aria-label={`SKU ${i + 1}`}
              className="flex-1 rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
            />
            <input
              type="number"
              min={1}
              value={l.qty}
              onChange={(e) => update(i, { qty: Number(e.target.value) })}
              aria-label={`Quantity ${i + 1}`}
              className="w-16 rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
            />
            {lines.length > 1 && (
              <button
                type="button"
                aria-label={`Remove line ${i + 1}`}
                onClick={() => setLines((cur) => cur.filter((_, idx) => idx !== i))}
                className="rounded-md px-2 py-1 text-sm text-muted-foreground hover:text-destructive"
              >
                ×
              </button>
            )}
          </div>
        ))}
      </div>
      <button
        type="button"
        onClick={() => setLines((cur) => [...cur, { sku: "", qty: 1 }])}
        className="text-xs font-medium text-primary"
      >
        + Add line
      </button>

      <div className="flex justify-end gap-2 pt-1">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground"
        >
          Cancel
        </button>
        <button
          type="button"
          disabled={create.isPending}
          onClick={submit}
          className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
        >
          {create.isPending ? "Creating…" : "Create"}
        </button>
      </div>
    </div>
  );
}
