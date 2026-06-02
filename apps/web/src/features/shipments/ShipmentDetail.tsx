import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import type { ShipmentBoard } from "./types";
import {
  shipmentStatusLabel,
  useCancelShipment,
  useDispatchShipment,
  usePackShipment,
} from "./useShipments";

export function ShipmentDetail({ storeId, board }: { storeId: string; board: ShipmentBoard }) {
  const toast = useToast();
  const pack = usePackShipment(storeId, board.id);
  const dispatch = useDispatchShipment(storeId, board.id);
  const cancel = useCancelShipment(storeId, board.id);

  const [carrier, setCarrier] = useState("");
  const [tracking, setTracking] = useState("");

  const total = board.items.reduce((n, i) => n + i.qty, 0);
  const open = board.status === "packing" || board.status === "packed";

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <h2 className="text-base font-semibold text-foreground">
            {board.reference || "Shipment"}
          </h2>
          <div className="text-xs text-muted-foreground">
            {shipmentStatusLabel(board.status)} · {board.items.length} SKUs · {total} units
          </div>
        </div>
        <div className="flex items-center gap-2">
          {board.status === "packing" && (
            <button
              type="button"
              disabled={pack.isPending}
              onClick={() =>
                pack.mutate(undefined, {
                  onSuccess: () => toast.success("Marked packed"),
                  onError: () => toast.error("Failed to pack"),
                })
              }
              className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
            >
              Mark packed
            </button>
          )}
          {open && (
            <button
              type="button"
              disabled={cancel.isPending}
              onClick={() =>
                cancel.mutate(undefined, {
                  onSuccess: () => toast.success("Shipment cancelled"),
                  onError: () => toast.error("Failed to cancel"),
                })
              }
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground disabled:opacity-50"
            >
              Cancel
            </button>
          )}
        </div>
      </header>

      <div className="flex-1 space-y-4 overflow-auto p-4">
        {board.status === "packed" && (
          <div className="space-y-2 rounded-lg border border-border bg-card p-3">
            <div className="text-xs font-medium text-muted-foreground">Dispatch</div>
            <div className="flex flex-wrap items-center gap-2">
              <input
                value={carrier}
                onChange={(e) => setCarrier(e.target.value)}
                placeholder="Carrier (e.g. UPS)"
                aria-label="Carrier"
                className="flex-1 rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
              />
              <input
                value={tracking}
                onChange={(e) => setTracking(e.target.value)}
                placeholder="Tracking number"
                aria-label="Tracking number"
                className="flex-1 rounded-md border border-border bg-surface px-2 py-1.5 text-sm text-foreground"
              />
              <button
                type="button"
                disabled={dispatch.isPending || !carrier.trim() || !tracking.trim()}
                onClick={() =>
                  dispatch.mutate(
                    { carrier: carrier.trim(), tracking_number: tracking.trim() },
                    {
                      onSuccess: () => toast.success("Dispatched"),
                      onError: () => toast.error("Failed to dispatch"),
                    },
                  )
                }
                className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
              >
                Dispatch
              </button>
            </div>
          </div>
        )}

        {board.status === "dispatched" && (
          <div className="rounded-lg border border-emerald-500/40 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-700">
            Dispatched via {board.carrier} · {board.tracking_number}
          </div>
        )}

        <ul className="rounded-lg border border-border bg-card">
          {board.items.map((it) => (
            <li
              key={it.sku}
              className="flex items-center justify-between border-b border-border px-3 py-2 text-sm last:border-0"
            >
              <span className="font-medium text-foreground">{it.sku}</span>
              <span className="text-muted-foreground">{it.qty}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
