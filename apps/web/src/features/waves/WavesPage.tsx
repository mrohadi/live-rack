import { useEffect, useState } from "react";
import { useCurrentStore } from "../map/useCurrentStore";
import { NewWaveForm } from "./NewWaveForm";
import { WaveRunView } from "./WaveRunView";
import type { WaveSummary } from "./types";
import { useWaveBoard, useWaves } from "./useWaves";

function statusDot(status: WaveSummary["status"]): string {
  switch (status) {
    case "completed":
      return "bg-emerald-500";
    case "picking":
      return "bg-amber-500";
    case "cancelled":
      return "bg-muted-foreground";
    default:
      return "bg-primary";
  }
}

export function WavesPage() {
  const storeId = useCurrentStore();
  const { data: waves = [], isLoading } = useWaves(storeId);
  const [selected, setSelected] = useState<string | undefined>();
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (!selected && !creating && waves.length > 0) setSelected(waves[0].id);
  }, [waves, selected, creating]);

  const { data: board } = useWaveBoard(storeId, creating ? undefined : selected);

  return (
    <div className="flex h-full">
      <aside className="flex w-72 shrink-0 flex-col border-r border-border">
        <header className="flex items-center justify-between border-b border-border px-4 py-3">
          <h1 className="text-lg font-semibold text-foreground">Waves</h1>
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
          ) : waves.length === 0 ? (
            <div className="p-2 text-sm text-muted-foreground">No waves yet.</div>
          ) : (
            <ul className="space-y-1">
              {waves.map((w) => (
                <li key={w.id}>
                  <button
                    type="button"
                    onClick={() => {
                      setSelected(w.id);
                      setCreating(false);
                    }}
                    className={`flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-left text-sm ${
                      selected === w.id && !creating
                        ? "bg-primary/10 text-foreground"
                        : "text-muted-foreground hover:bg-muted"
                    }`}
                  >
                    <span className={`h-2 w-2 shrink-0 rounded-full ${statusDot(w.status)}`} />
                    <span className="min-w-0 flex-1 truncate">{w.reference || "Untitled"}</span>
                    <span className="shrink-0 text-xs text-muted-foreground">
                      {w.list_count} lists
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
            <NewWaveForm
              storeId={storeId}
              onCreated={(id) => {
                setCreating(false);
                setSelected(id);
              }}
              onCancel={() => setCreating(false)}
            />
          </div>
        ) : board ? (
          <WaveRunView storeId={storeId} board={board} />
        ) : (
          <div className="p-6 text-sm text-muted-foreground">
            Select a wave or create a new one.
          </div>
        )}
      </main>
    </div>
  );
}
