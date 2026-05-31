import { useDraggable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import type { Card, CardPriority } from "./types";
import { formatAge } from "./usePipelines";

const PRIORITY_STYLES: Record<CardPriority, string> = {
  high: "bg-destructive/15 text-destructive",
  medium: "bg-warning/15 text-warning",
  low: "bg-muted/40 text-muted-foreground",
};

export function PipelineCard({ card }: { card: Card }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: card.id,
  });

  return (
    <div
      ref={setNodeRef}
      data-testid="pipeline-card"
      data-card-id={card.id}
      style={{ transform: CSS.Translate.toString(transform) }}
      className={`cursor-grab rounded-lg border bg-surface p-3 shadow-sm ${
        card.ageing ? "border-destructive/50 ring-1 ring-destructive/30" : "border-border"
      } ${isDragging ? "opacity-50" : ""}`}
      {...listeners}
      {...attributes}
    >
      <div className="mb-1 flex items-center gap-2 text-[11px] text-muted-foreground">
        <span className="font-mono">{card.id.slice(0, 8)}</span>
        {card.sku && <span>· {card.sku}</span>}
      </div>
      <div className="mb-2 text-sm font-medium text-foreground">{card.title}</div>
      <div className="flex items-center justify-between">
        <span className={`rounded px-1.5 py-0.5 text-xs ${PRIORITY_STYLES[card.priority]}`}>
          {card.priority}
        </span>
        <span
          className={`font-mono text-[11px] ${
            card.ageing ? "font-semibold text-destructive" : "text-muted-foreground"
          }`}
          title={card.ageing ? "Past stage SLA" : "Within SLA"}
        >
          {formatAge(card.age_seconds)}
        </span>
      </div>
    </div>
  );
}
