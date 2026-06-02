import { useDraggable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import type { Task, TaskPriority } from "./types";

const PRIORITY_STYLES: Record<TaskPriority, string> = {
  high: "bg-destructive/15 text-destructive",
  med: "bg-warning/15 text-warning",
  low: "bg-muted/40 text-muted-foreground",
};

export function TaskCard({ task }: { task: Task }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: task.id,
  });

  return (
    <div
      ref={setNodeRef}
      data-testid="task-card"
      data-task-id={task.id}
      style={{ transform: CSS.Translate.toString(transform) }}
      className={`cursor-grab rounded-lg border border-border bg-surface p-3 shadow-sm ${
        isDragging ? "opacity-50" : ""
      }`}
      {...listeners}
      {...attributes}
    >
      <div className="mb-1 flex items-center gap-2 text-[11px] text-muted-foreground">
        <span className="font-mono">{task.id.slice(0, 8)}</span>
        {task.zone_id && <span>· {task.zone_id.slice(0, 8)}</span>}
      </div>
      <div className="mb-2 text-sm font-medium text-foreground">{task.title}</div>
      <div className="flex items-center justify-between">
        <span className={`rounded px-1.5 py-0.5 text-xs ${PRIORITY_STYLES[task.priority]}`}>
          {task.priority}
        </span>
        {task.due_at && (
          <span className="text-[11px] text-muted-foreground">
            {new Date(task.due_at).toLocaleDateString()}
          </span>
        )}
      </div>
    </div>
  );
}
