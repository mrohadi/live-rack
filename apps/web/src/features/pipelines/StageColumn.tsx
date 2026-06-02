import { useDroppable } from "@dnd-kit/core";
import { PipelineCard } from "./PipelineCard";
import type { Card } from "./types";

export function StageColumn({
  position,
  name,
  cards,
}: {
  position: number;
  name: string;
  cards: Card[];
}) {
  const { setNodeRef, isOver } = useDroppable({ id: String(position) });

  return (
    <div
      ref={setNodeRef}
      data-testid={`stage-${position}`}
      className={`flex min-h-64 flex-col gap-2 rounded-lg border border-border bg-muted/20 p-3 ${
        isOver ? "ring-2 ring-primary" : ""
      }`}
    >
      <div className="flex items-center justify-between text-sm font-semibold text-foreground">
        <span>
          {position + 1}. {name}
        </span>
        <span className="rounded-full bg-muted px-2 text-xs text-muted-foreground">
          {cards.length}
        </span>
      </div>
      {cards.map((c) => (
        <PipelineCard key={c.id} card={c} />
      ))}
    </div>
  );
}
