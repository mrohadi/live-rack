/** Kanban column — mirrors the tasks.status CHECK constraint. */
export type TaskStatus = "todo" | "in_progress" | "review" | "done";

/** Card priority — orders cards within a column. */
export type TaskPriority = "low" | "med" | "high";

export interface Task {
  id: string;
  store_id: string;
  zone_id?: string;
  title: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee_id?: string;
  due_at?: string;
  updated_at: string;
}

/** Ordered kanban columns with display labels. */
export const TASK_COLUMNS: { status: TaskStatus; label: string }[] = [
  { status: "todo", label: "To do" },
  { status: "in_progress", label: "In progress" },
  { status: "review", label: "Review" },
  { status: "done", label: "Done" },
];
