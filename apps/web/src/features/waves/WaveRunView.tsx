import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { PickRoute } from "../picking/PickRoute";
import { lineStatusLabel } from "../picking/usePicking";
import type { WaveBoard, WaveStop } from "./types";
import {
  stopToLine,
  useCompleteWave,
  useConfirmStop,
  useStartWave,
  waveProgress,
} from "./useWaves";

function StatusPill({ status }: { status: WaveStop["status"] }) {
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

function StopRow({
  stop,
  disabled,
  onConfirm,
}: {
  stop: WaveStop;
  disabled: boolean;
  onConfirm: (qty: number) => void;
}) {
  const [qty, setQty] = useState(stop.qty_requested);
  const pending = stop.status === "pending";

  return (
    <li className="flex items-center gap-3 border-b border-border px-3 py-2 last:border-0">
      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
        {stop.seq + 1}
      </span>
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium text-foreground">{stop.sku}</div>
        <div className="truncate text-xs text-muted-foreground">
          {stop.zone_name || "Unmapped"} · {stop.order_count} orders · need {stop.qty_requested}
        </div>
      </div>
      {pending ? (
        <div className="flex items-center gap-2">
          <input
            type="number"
            min={0}
            max={stop.qty_requested}
            value={qty}
            onChange={(e) => setQty(Number(e.target.value))}
            aria-label={`Quantity picked for ${stop.sku}`}
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
            {stop.qty_picked}/{stop.qty_requested}
          </span>
          <StatusPill status={stop.status} />
        </div>
      )}
    </li>
  );
}

export function WaveRunView({ storeId, board }: { storeId: string; board: WaveBoard }) {
  const toast = useToast();
  const start = useStartWave(storeId, board.id);
  const confirm = useConfirmStop(storeId, board.id);
  const complete = useCompleteWave(storeId, board.id);

  const { done, total, pct } = waveProgress(board.stops);
  const finished = board.status === "completed" || board.status === "cancelled";

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <h2 className="text-base font-semibold text-foreground">{board.reference || "Wave"}</h2>
          <div className="text-xs capitalize text-muted-foreground">{board.status}</div>
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
                  onSuccess: () => toast.success("Wave completed"),
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
              {done}/{total} stops
            </span>
            <span>{pct}%</span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
            <div className="h-full bg-primary transition-all" style={{ width: `${pct}%` }} />
          </div>
        </div>

        <PickRoute lines={board.stops.map(stopToLine)} />

        <ul className="rounded-lg border border-border bg-card">
          {board.stops.map((s) => (
            <StopRow
              key={`${s.sku}|${s.zone_id}`}
              stop={s}
              disabled={confirm.isPending || finished || board.status === "open"}
              onConfirm={(qty) =>
                confirm.mutate(
                  { sku: s.sku, zone_id: s.zone_id ?? "", qty_picked: qty },
                  {
                    onSuccess: () =>
                      qty < s.qty_requested
                        ? toast.error(`Short — restock raised for ${s.sku}`)
                        : toast.success(`Picked ${s.sku}`),
                    onError: () => toast.error("Failed to confirm"),
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
