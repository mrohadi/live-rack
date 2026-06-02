import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { Select } from "../../components/ui/Select";
import { useMembers } from "../users/useUsers";
import { useAssignTask } from "./useTasks";
import { useCurrentStore } from "../map/useCurrentStore";
import type { Task } from "./types";

interface Props {
  task: Task;
  onClose: () => void;
}

export function AssignMemberModal({ task, onClose }: Props) {
  const storeId = useCurrentStore();
  const assignTask = useAssignTask(storeId);
  const { data: members = [], isLoading } = useMembers();
  const toast = useToast();

  const [assigneeId, setAssigneeId] = useState(task.assignee_id ?? "");

  const memberOptions = members.map((m) => ({
    value: m.id,
    label: m.display_name || m.email,
    sub: m.email !== (m.display_name || m.email) ? m.email : undefined,
    avatar: (m.display_name || m.email)[0].toUpperCase(),
  }));

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!assigneeId) return;
    assignTask.mutate(
      { taskId: task.id, assigneeId },
      {
        onSuccess: () => {
          toast.success("Member assigned");
          onClose();
        },
        onError: () => toast.error("Failed to assign member"),
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Assign member"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <form
        onSubmit={submit}
        className="w-full max-w-sm space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-foreground">Assign member</h2>
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
          <span className="text-muted-foreground">Task</span>
          <span className="font-medium text-foreground">{task.title}</span>
        </div>

        <label className="block text-sm">
          <span className="mb-1.5 block text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Select member *
          </span>
          <Select
            value={assigneeId}
            onChange={setAssigneeId}
            options={memberOptions}
            placeholder={isLoading ? "Loading…" : "— pick a member —"}
            searchable={members.length > 5}
            disabled={isLoading}
          />
        </label>

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
            disabled={assignTask.isPending || !assigneeId}
            className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
          >
            {assignTask.isPending ? "Saving…" : "Assign"}
          </button>
        </div>
      </form>
    </div>
  );
}
