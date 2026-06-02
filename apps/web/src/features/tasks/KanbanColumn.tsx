import { useDroppable } from "@dnd-kit/core";
import { TaskCard } from "./TaskCard";
import type { Task, TaskStatus } from "./types";
import type { OrgUser } from "../users/useUsers";

export function KanbanColumn({
  status,
  label,
  tasks,
  members,
  onAssign,
}: {
  status: TaskStatus;
  label: string;
  tasks: Task[];
  members?: OrgUser[];
  onAssign?: (task: Task) => void;
}) {
  const { setNodeRef, isOver } = useDroppable({ id: status });

  return (
    <div
      ref={setNodeRef}
      data-testid={`column-${status}`}
      className={`flex min-h-64 flex-col gap-2 rounded-lg border border-border bg-muted/20 p-3 ${
        isOver ? "ring-2 ring-primary" : ""
      }`}
    >
      <div className="flex items-center justify-between text-sm font-semibold text-foreground">
        <span>{label}</span>
        <span className="rounded-full bg-muted px-2 text-xs text-muted-foreground">
          {tasks.length}
        </span>
      </div>
      {tasks.map((t) => (
        <TaskCard key={t.id} task={t} members={members} onAssign={onAssign} />
      ))}
    </div>
  );
}
