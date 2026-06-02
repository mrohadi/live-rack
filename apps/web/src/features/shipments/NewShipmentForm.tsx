import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { usePickLists } from "../picking/usePicking";
import { useCreateShipment } from "./useShipments";

interface Props {
  storeId: string;
  onCreated: (id: string) => void;
  onCancel: () => void;
}

export function NewShipmentForm({ storeId, onCreated, onCancel }: Props) {
  const toast = useToast();
  const { data: lists = [], isLoading } = usePickLists(storeId);
  const create = useCreateShipment(storeId);
  const [reference, setReference] = useState("");
  const [pickListId, setPickListId] = useState("");

  // Only completed pick lists can be packed into a shipment.
  const eligible = lists.filter((l) => l.status === "completed");

  function submit() {
    if (!pickListId) {
      toast.error("Select a completed pick list");
      return;
    }
    create.mutate(
      { pick_list_id: pickListId, reference: reference.trim() },
      {
        onSuccess: (board) => {
          toast.success("Shipment created");
          onCreated(board.id);
        },
        onError: () => toast.error("Failed to create shipment"),
      },
    );
  }

  return (
    <div className="space-y-3 rounded-lg border border-border bg-card p-3">
      <input
        value={reference}
        onChange={(e) => setReference(e.target.value)}
        placeholder="Shipment reference (e.g. SHIP-204)"
        aria-label="Shipment reference"
        className="w-full rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
      />
      <select
        value={pickListId}
        onChange={(e) => setPickListId(e.target.value)}
        aria-label="Completed pick list"
        className="w-full rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
      >
        <option value="">{isLoading ? "Loading…" : "Select a completed pick list…"}</option>
        {eligible.map((l) => (
          <option key={l.id} value={l.id}>
            {l.reference || "Untitled"} · {l.line_count} lines
          </option>
        ))}
      </select>
      {!isLoading && eligible.length === 0 && (
        <div className="text-xs text-muted-foreground">No completed pick lists to pack.</div>
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
          disabled={create.isPending || !pickListId}
          onClick={submit}
          className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
        >
          {create.isPending ? "Creating…" : "Create shipment"}
        </button>
      </div>
    </div>
  );
}
