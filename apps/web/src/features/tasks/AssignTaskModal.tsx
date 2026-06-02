import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { useCreateTask } from "./useTasks";
import { useCurrentStore } from "../map/useCurrentStore";
import type { TaskPriority } from "./types";

interface Props {
  zoneId: string;
  zoneName: string;
  onClose: () => void;
}

const PRIORITIES: { value: TaskPriority; label: string }[] = [
  { value: "low", label: "Low" },
  { value: "med", label: "Medium" },
  { value: "high", label: "High" },
];

export function AssignTaskModal({ zoneId, zoneName, onClose }: Props) {
  const storeId = useCurrentStore();
  const createTask = useCreateTask(storeId);
  const toast = useToast();

  const [title, setTitle] = useState("");
  const [priority, setPriority] = useState<TaskPriority>("med");
  const [dueAt, setDueAt] = useState("");

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    createTask.mutate(
      {
        zone_id: zoneId,
        title: title.trim(),
        priority,
        due_at: dueAt ? new Date(dueAt).toISOString() : undefined,
      },
      {
        onSuccess: () => {
          toast.success("Task created");
          onClose();
        },
        onError: () => toast.error("Failed to create task"),
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Assign task"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <form
        onSubmit={submit}
        className="w-full max-w-sm space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-foreground">Assign task</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            ✕
          </button>
        </div>

        <div className="rounded-md bg-muted px-3 py-2 text-sm text-muted-foreground">
          Zone: <span className="font-medium text-foreground">{zoneName}</span>
        </div>

        <Field label="Task title *">
          <input
            autoFocus
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Restock frozen section"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <Field label="Priority">
          <select
            value={priority}
            onChange={(e) => setPriority(e.target.value as TaskPriority)}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          >
            {PRIORITIES.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </select>
        </Field>

        <Field label="Due date (optional)">
          <input
            type="date"
            value={dueAt}
            onChange={(e) => setDueAt(e.target.value)}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </Field>

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground hover:bg-muted"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createTask.isPending}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            {createTask.isPending ? "Creating…" : "Create task"}
          </button>
        </div>
      </form>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block text-sm">
      <span className="mb-1 block text-muted-foreground">{label}</span>
      {children}
    </label>
  );
}
