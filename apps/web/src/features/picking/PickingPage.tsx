import { useEffect, useState } from "react";
import { useCurrentStore } from "../map/useCurrentStore";
import { NewPickListForm } from "./NewPickListForm";
import { PickRunView } from "./PickRunView";
import type { PickListSummary } from "./types";
import { usePickBoard, usePickLists } from "./usePicking";

function statusDot(status: PickListSummary["status"]): string {
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

export function PickingPage() {
  const storeId = useCurrentStore();
  const { data: lists = [], isLoading } = usePickLists(storeId);
  const [selected, setSelected] = useState<string | undefined>();
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (!selected && !creating && lists.length > 0) setSelected(lists[0].id);
  }, [lists, selected, creating]);

  const { data: board } = usePickBoard(storeId, creating ? undefined : selected);

  return (
    <div className="flex h-full">
      <aside className="flex w-72 shrink-0 flex-col border-r border-border">
        <header className="flex items-center justify-between border-b border-border px-4 py-3">
          <h1 className="text-lg font-semibold text-foreground">Picking</h1>
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
          ) : lists.length === 0 ? (
            <div className="p-2 text-sm text-muted-foreground">No pick lists yet.</div>
          ) : (
            <ul className="space-y-1">
              {lists.map((l) => (
                <li key={l.id}>
                  <button
                    type="button"
                    onClick={() => {
                      setSelected(l.id);
                      setCreating(false);
                    }}
                    className={`flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-left text-sm ${
                      selected === l.id && !creating
                        ? "bg-primary/10 text-foreground"
                        : "text-muted-foreground hover:bg-muted"
                    }`}
                  >
                    <span className={`h-2 w-2 shrink-0 rounded-full ${statusDot(l.status)}`} />
                    <span className="min-w-0 flex-1 truncate">{l.reference || "Untitled"}</span>
                    <span className="shrink-0 text-xs text-muted-foreground">
                      {l.done_count}/{l.line_count}
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
            <NewPickListForm
              storeId={storeId}
              onCreated={(id) => {
                setCreating(false);
                setSelected(id);
              }}
              onCancel={() => setCreating(false)}
            />
          </div>
        ) : board ? (
          <PickRunView storeId={storeId} board={board} />
        ) : (
          <div className="p-6 text-sm text-muted-foreground">
            Select a pick list or create a new one.
          </div>
        )}
      </main>
    </div>
  );
}
