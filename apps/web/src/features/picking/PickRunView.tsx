import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { PickRoute } from "./PickRoute";
import type { PickBoard, PickLine } from "./types";
import {
  lineStatusLabel,
  pickProgress,
  useCompletePick,
  useConfirmPick,
  useStartPick,
} from "./usePicking";

function StatusPill({ status }: { status: PickLine["status"] }) {
  const cls =
    status === "picked"
      ? "bg-emerald-500/15 text-emerald-600"
      : status === "short"
        ? "bg-destructive/15 text-destructive"
        : "bg-muted text-muted-foreground";
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {lineStatusLabel(status)}
    </span>
  );
}

function PickLineRow({
  line,
  disabled,
  onConfirm,
}: {
  line: PickLine;
  disabled: boolean;
  onConfirm: (qty: number) => void;
}) {
  const [qty, setQty] = useState(line.qty_requested);
  const pending = line.status === "pending";

  return (
    <li className="flex items-center gap-3 border-b border-border px-3 py-2 last:border-0">
      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
        {line.seq + 1}
      </span>
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium text-foreground">{line.sku}</div>
        <div className="truncate text-xs text-muted-foreground">
          {line.zone_name || "Unmapped"} · need {line.qty_requested}
        </div>
      </div>
      {pending ? (
        <div className="flex items-center gap-2">
          <input
            type="number"
            min={0}
            max={line.qty_requested}
            value={qty}
            onChange={(e) => setQty(Number(e.target.value))}
            aria-label={`Quantity picked for ${line.sku}`}
            className="w-16 rounded-md border border-border bg-surface px-2 py-1 text-sm text-foreground"
          />
          <button
            type="button"
            disabled={disabled}
            onClick={() => onConfirm(qty)}
            className="rounded-md bg-primary px-2.5 py-1 text-xs font-medium text-white disabled:opacity-50"
          >
            Confirm
          </button>
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground">
            {line.qty_picked}/{line.qty_requested}
          </span>
          <StatusPill status={line.status} />
        </div>
      )}
    </li>
  );
}

export function PickRunView({ storeId, board }: { storeId: string; board: PickBoard }) {
  const toast = useToast();
  const start = useStartPick(storeId, board.id);
  const confirm = useConfirmPick(storeId, board.id);
  const complete = useCompletePick(storeId, board.id);

  const { done, total, pct } = pickProgress(board.lines);
  const finished = board.status === "completed" || board.status === "cancelled";

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <h2 className="text-base font-semibold text-foreground">
            {board.reference || "Pick list"}
          </h2>
          <div className="text-xs text-muted-foreground capitalize">{board.status}</div>
        </div>
        <div className="flex items-center gap-2">
          {board.status === "open" && (
            <button
              type="button"
              disabled={start.isPending}
              onClick={() =>
                start.mutate(undefined, { onError: () => toast.error("Failed to start") })
              }
              className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
            >
              Start picking
            </button>
          )}
          {!finished && (
            <button
              type="button"
              disabled={complete.isPending}
              onClick={() =>
                complete.mutate(undefined, {
                  onSuccess: () => toast.success("Pick list completed"),
                  onError: () => toast.error("Failed to complete"),
                })
              }
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground disabled:opacity-50"
            >
              Complete
            </button>
          )}
        </div>
      </header>

      <div className="flex-1 space-y-4 overflow-auto p-4">
        <div>
          <div className="mb-1 flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {done}/{total} picked
            </span>
            <span>{pct}%</span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
            <div className="h-full bg-primary transition-all" style={{ width: `${pct}%` }} />
          </div>
        </div>

        <PickRoute lines={board.lines} />

        <ul className="rounded-lg border border-border bg-card">
          {board.lines.map((l) => (
            <PickLineRow
              key={l.id}
              line={l}
              disabled={confirm.isPending || finished || board.status === "open"}
              onConfirm={(qty) =>
                confirm.mutate(
                  { lineId: l.id, qtyPicked: qty },
                  {
                    onSuccess: () =>
                      qty < l.qty_requested
                        ? toast.error(`Short pick — restock task raised for ${l.sku}`)
                        : toast.success(`Picked ${l.sku}`),
                    onError: () => toast.error("Failed to confirm pick"),
                  },
                )
              }
            />
          ))}
        </ul>
      </div>
    </div>
  );
}
