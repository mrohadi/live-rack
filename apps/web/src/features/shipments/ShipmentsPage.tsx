import { useEffect, useState } from "react";
import { useCurrentStore } from "../map/useCurrentStore";
import { NewShipmentForm } from "./NewShipmentForm";
import { ShipmentDetail } from "./ShipmentDetail";
import type { ShipmentSummary } from "./types";
import { useShipmentBoard, useShipments } from "./useShipments";

function statusDot(status: ShipmentSummary["status"]): string {
  switch (status) {
    case "dispatched":
      return "bg-emerald-500";
    case "packed":
      return "bg-amber-500";
    case "cancelled":
      return "bg-muted-foreground";
    default:
      return "bg-primary";
  }
}

export function ShipmentsPage() {
  const storeId = useCurrentStore();
  const { data: shipments = [], isLoading } = useShipments(storeId);
  const [selected, setSelected] = useState<string | undefined>();
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (!selected && !creating && shipments.length > 0) setSelected(shipments[0].id);
  }, [shipments, selected, creating]);

  const { data: board } = useShipmentBoard(storeId, creating ? undefined : selected);

  return (
    <div className="flex h-full">
      <aside className="flex w-72 shrink-0 flex-col border-r border-border">
        <header className="flex items-center justify-between border-b border-border px-4 py-3">
          <h1 className="text-lg font-semibold text-foreground">Dispatch</h1>
          <button
            type="button"
            onClick={() => {
              setCreating(true);
              setSelected(undefined);
            }}
            className="rounded-md bg-primary px-2.5 py-1 text-xs font-medium text-white"
          >
            + New
          </button>
        </header>
        <div className="flex-1 overflow-auto p-2">
          {isLoading ? (
            <div className="p-2 text-sm text-muted-foreground">Loading…</div>
          ) : shipments.length === 0 ? (
            <div className="p-2 text-sm text-muted-foreground">No shipments yet.</div>
          ) : (
            <ul className="space-y-1">
              {shipments.map((s) => (
                <li key={s.id}>
                  <button
                    type="button"
                    onClick={() => {
                      setSelected(s.id);
                      setCreating(false);
                    }}
                    className={`flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-left text-sm ${
                      selected === s.id && !creating
                        ? "bg-primary/10 text-foreground"
                        : "text-muted-foreground hover:bg-muted"
                    }`}
                  >
                    <span className={`h-2 w-2 shrink-0 rounded-full ${statusDot(s.status)}`} />
                    <span className="min-w-0 flex-1 truncate">{s.reference || "Untitled"}</span>
                    <span className="shrink-0 text-xs text-muted-foreground">
                      {s.item_count} SKUs
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </aside>

      <main className="min-w-0 flex-1">
        {creating ? (
          <div className="p-4">
            <NewShipmentForm
              storeId={storeId}
              onCreated={(id) => {
                setCreating(false);
                setSelected(id);
              }}
              onCancel={() => setCreating(false)}
            />
          </div>
        ) : board ? (
          <ShipmentDetail storeId={storeId} board={board} />
        ) : (
          <div className="p-6 text-sm text-muted-foreground">
            Select a shipment or create a new one.
          </div>
        )}
      </main>
    </div>
  );
}
