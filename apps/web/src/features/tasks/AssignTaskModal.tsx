import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { Select } from "../../components/ui/Select";
import { DatePicker } from "../../components/ui/DatePicker";
import { useMembers } from "../users/useUsers";
import { useCreateTask } from "./useTasks";
import { useCurrentStore } from "../map/useCurrentStore";
import type { TaskPriority } from "./types";

interface Props {
  zoneId: string;
  zoneName: string;
  onClose: () => void;
}

const PRIORITY_OPTIONS = [
  {
    value: "low",
    label: "Low",
    badge: "low",
    badgeClass: "bg-muted/40 text-muted-foreground",
  },
  {
    value: "med",
    label: "Medium",
    badge: "med",
    badgeClass: "bg-warning/15 text-warning",
  },
  {
    value: "high",
    label: "High",
    badge: "high",
    badgeClass: "bg-destructive/15 text-destructive",
  },
];

export function AssignTaskModal({ zoneId, zoneName, onClose }: Props) {
  const storeId = useCurrentStore();
  const createTask = useCreateTask(storeId);
  const { data: members = [] } = useMembers();
  const toast = useToast();

  const [title, setTitle] = useState("");
  const [priority, setPriority] = useState<TaskPriority>("med");
  const [dueAt, setDueAt] = useState("");
  const [assigneeId, setAssigneeId] = useState("");

  const memberOptions = [
    { value: "", label: "Unassigned", sub: "no assignee" },
    ...members.map((m) => ({
      value: m.id,
      label: m.display_name || m.email,
      sub: m.email !== (m.display_name || m.email) ? m.email : undefined,
      avatar: ((m.display_name || m.email || "?")[0] ?? "?").toUpperCase(),
    })),
  ];

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    createTask.mutate(
      {
        zone_id: zoneId,
        title: title.trim(),
        priority,
        due_at: dueAt ? new Date(dueAt).toISOString() : undefined,
        assignee_id: assigneeId || undefined,
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

        <div className="flex items-center gap-2 rounded-lg bg-muted px-3 py-2 text-sm">
          <span className="text-muted-foreground">Zone</span>
          <span className="font-medium text-foreground">{zoneName}</span>
        </div>

        <Field label="Task title *">
          <input
            autoFocus
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Restock frozen section"
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
          />
        </Field>

        <Field label="Priority">
          <Select
            value={priority}
            onChange={(v) => setPriority(v as TaskPriority)}
            options={PRIORITY_OPTIONS}
          />
        </Field>

        <Field label="Assign to (optional)">
          <Select
            value={assigneeId}
            onChange={setAssigneeId}
            options={memberOptions}
            placeholder="Unassigned"
            searchable={members.length > 5}
          />
        </Field>

        <Field label="Due date (optional)">
          <DatePicker value={dueAt} onChange={setDueAt} />
        </Field>

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-border px-3 py-1.5 text-sm text-foreground transition hover:bg-muted"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createTask.isPending}
            className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
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
      <span className="mb-1.5 block text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </span>
      {children}
    </label>
  );
}
