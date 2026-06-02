import { DndContext, PointerSensor, useSensor, useSensors, type DragEndEvent } from "@dnd-kit/core";
import { useEffect, useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { useCurrentStore } from "../map/useCurrentStore";
import { StageColumn } from "./StageColumn";
import {
  ageingCount,
  cardsByStage,
  useBoard,
  useCreateFromTemplate,
  useMoveCard,
  usePipelines,
} from "./usePipelines";

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
  const createFromTemplate = useCreateFromTemplate(storeId);
  const toast = useToast();

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));

  function onDragEnd(ev: DragEndEvent) {
    if (!board) return;
    const target = ev.over?.id;
    if (target == null) return;
    const stagePosition = Number(target);
    const id = String(ev.active.id);
    const card = board.cards.find((c) => c.id === id);
    if (!card || card.stage_position === stagePosition) return;
    move.mutate({ id, stagePosition }, { onError: () => toast.error("Failed to move card") });
  }

  if (loadingList) {
    return <div className="p-6 text-sm text-muted-foreground">Loading pipelines…</div>;
  }
  if (pipelines.length === 0) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-sm text-muted-foreground">
        <span>No pipelines yet.</span>
        <button
          type="button"
          disabled={createFromTemplate.isPending}
          onClick={() =>
            createFromTemplate.mutate("item-restoration", {
              onSuccess: () => toast.success("Pipeline created"),
              onError: () => toast.error("Failed to create pipeline"),
            })
          }
          className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
        >
          {createFromTemplate.isPending ? "Creating…" : "Start Item Restoration template"}
        </button>
      </div>
    );
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
        {board?.bottleneck && (
          <div
            role="alert"
            className="mb-3 flex items-center gap-2 rounded-lg border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            <span className="font-semibold">Bottleneck · {board.bottleneck.name}</span>
            <span className="text-destructive/80">
              {board.bottleneck.ageing_count} card
              {board.bottleneck.ageing_count === 1 ? "" : "s"} past SLA
            </span>
          </div>
        )}
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
