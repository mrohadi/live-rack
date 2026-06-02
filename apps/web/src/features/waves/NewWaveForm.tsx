import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { usePickLists } from "../picking/usePicking";
import { useCreateWave } from "./useWaves";

interface Props {
  storeId: string;
  onCreated: (id: string) => void;
  onCancel: () => void;
}

export function NewWaveForm({ storeId, onCreated, onCancel }: Props) {
  const toast = useToast();
  const { data: lists = [], isLoading } = usePickLists(storeId);
  const create = useCreateWave(storeId);
  const [reference, setReference] = useState("");
  const [picked, setPicked] = useState<Record<string, boolean>>({});

  // Only unstarted lists can join a wave.
  const eligible = lists.filter((l) => l.status === "open");
  const selectedIds = Object.keys(picked).filter((id) => picked[id]);

  function submit() {
    if (selectedIds.length < 2) {
      toast.error("Select at least two pick lists");
      return;
    }
    create.mutate(
      { reference: reference.trim(), list_ids: selectedIds },
      {
        onSuccess: (board) => {
          toast.success("Wave created");
          onCreated(board.id);
        },
        onError: () => toast.error("Failed to create wave"),
      },
    );
  }

  return (
    <div className="space-y-3 rounded-lg border border-border bg-card p-3">
      <input
        value={reference}
        onChange={(e) => setReference(e.target.value)}
        placeholder="Wave reference (e.g. WAVE-12)"
        aria-label="Wave reference"
        className="w-full rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
      />
      <div className="text-xs font-medium text-muted-foreground">Pick lists to batch</div>
      {isLoading ? (
        <div className="text-sm text-muted-foreground">Loading…</div>
      ) : eligible.length === 0 ? (
        <div className="text-sm text-muted-foreground">No open pick lists to batch.</div>
      ) : (
        <ul className="max-h-56 space-y-1 overflow-auto">
          {eligible.map((l) => (
            <li key={l.id}>
              <label className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted">
                <input
                  type="checkbox"
                  checked={Boolean(picked[l.id])}
                  onChange={(e) => setPicked((cur) => ({ ...cur, [l.id]: e.target.checked }))}
                />
                <span className="min-w-0 flex-1 truncate text-foreground">
                  {l.reference || "Untitled"}
                </span>
                <span className="text-xs text-muted-foreground">{l.line_count} lines</span>
              </label>
            </li>
          ))}
        </ul>
      )}

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
          disabled={create.isPending || selectedIds.length < 2}
          onClick={submit}
          className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
        >
          {create.isPending ? "Creating…" : `Create wave (${selectedIds.length})`}
        </button>
      </div>
    </div>
  );
}
