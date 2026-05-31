import { DndContext, PointerSensor, useSensor, useSensors, type DragEndEvent } from "@dnd-kit/core";
import { useEffect, useState } from "react";
import { useCurrentStore } from "../map/useCurrentStore";
import { StageColumn } from "./StageColumn";
import { ageingCount, cardsByStage, useBoard, useMoveCard, usePipelines } from "./usePipelines";

export function PipelinesPage() {
  const storeId = useCurrentStore();
  const { data: pipelines = [], isLoading: loadingList } = usePipelines(storeId);
  const [selected, setSelected] = useState<string | undefined>();

  // Default to the first pipeline once the list loads.
  useEffect(() => {
    if (!selected && pipelines.length > 0) setSelected(pipelines[0].id);
  }, [pipelines, selected]);

  const { data: board, isLoading: loadingBoard } = useBoard(storeId, selected);
  const move = useMoveCard(storeId, selected ?? "");

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));

  function onDragEnd(ev: DragEndEvent) {
    if (!board) return;
    const target = ev.over?.id;
    if (target == null) return;
    const stagePosition = Number(target);
    const id = String(ev.active.id);
    const card = board.cards.find((c) => c.id === id);
    if (!card || card.stage_position === stagePosition) return;
    move.mutate({ id, stagePosition });
  }

  if (loadingList) {
    return <div className="p-6 text-sm text-muted-foreground">Loading pipelines…</div>;
  }
  if (pipelines.length === 0) {
    return <div className="p-6 text-sm text-muted-foreground">No pipelines yet.</div>;
  }

  const ageing = board ? ageingCount(board.cards) : 0;
  const columns = board
    ? cardsByStage(
        board.stages.map((s) => s.position),
        board.cards,
      )
    : {};

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-semibold text-foreground">Pipelines</h1>
          <select
            aria-label="Select pipeline"
            value={selected ?? ""}
            onChange={(e) => setSelected(e.target.value)}
            className="rounded-md border border-border bg-surface px-2 py-1 text-sm text-foreground"
          >
            {pipelines.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>
        {ageing > 0 && (
          <span className="rounded-full bg-destructive/15 px-2 py-0.5 text-xs font-medium text-destructive">
            {ageing} ageing · past SLA
          </span>
        )}
      </header>

      <div className="flex-1 overflow-auto p-4">
        {loadingBoard || !board ? (
          <div className="text-sm text-muted-foreground">Loading board…</div>
        ) : (
          <DndContext sensors={sensors} onDragEnd={onDragEnd}>
            <div className="grid grid-cols-1 gap-3 md:grid-cols-3 xl:grid-cols-5">
              {board.stages.map((s) => (
                <StageColumn
                  key={s.position}
                  position={s.position}
                  name={s.name}
                  cards={columns[s.position] ?? []}
                />
              ))}
            </div>
          </DndContext>
        )}
      </div>
    </div>
  );
}
