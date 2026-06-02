import { DndContext, PointerSensor, useSensor, useSensors, type DragEndEvent } from "@dnd-kit/core";
import { useToast } from "../../components/feedback/toast-context";
import { useCurrentStore } from "../map/useCurrentStore";
import { KanbanColumn } from "./KanbanColumn";
import { TASK_COLUMNS, type TaskStatus } from "./types";
import { groupByStatus, useMoveTask, useTasks } from "./useTasks";

export function TasksPage() {
  const storeId = useCurrentStore();
  const { data: tasks = [], isLoading } = useTasks(storeId);
  const move = useMoveTask(storeId);
  const toast = useToast();

  // Require a small drag distance so clicks don't trigger a move.
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));

  function onDragEnd(ev: DragEndEvent) {
    const status = ev.over?.id as TaskStatus | undefined;
    const id = String(ev.active.id);
    if (!status) return;
    const task = tasks.find((t) => t.id === id);
    if (!task || task.status === status) return;
    move.mutate({ id, status }, { onError: () => toast.error("Failed to move task") });
  }

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">Loading tasks…</div>;
  }

  const columns = groupByStatus(tasks);

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <h1 className="text-lg font-semibold text-foreground">Tasks</h1>
        <span className="text-xs text-muted-foreground">
          {tasks.length} open · drag cards between columns
        </span>
      </header>

      <div className="flex-1 overflow-auto p-4">
        <DndContext sensors={sensors} onDragEnd={onDragEnd}>
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
            {TASK_COLUMNS.map((col) => (
              <KanbanColumn
                key={col.status}
                status={col.status}
                label={col.label}
                tasks={columns[col.status]}
              />
            ))}
          </div>
        </DndContext>
      </div>
    </div>
  );
}
