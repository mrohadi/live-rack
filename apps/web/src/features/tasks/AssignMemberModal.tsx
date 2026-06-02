import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
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

  const currentMember = members.find((m) => m.id === task.assignee_id);

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

        <div className="rounded-md bg-muted px-3 py-2 text-sm text-muted-foreground">
          Task: <span className="font-medium text-foreground">{task.title}</span>
        </div>

        {currentMember && (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-primary/20 text-xs font-semibold text-primary">
              {(currentMember.display_name || currentMember.email)[0].toUpperCase()}
            </span>
            Currently: {currentMember.display_name || currentMember.email}
          </div>
        )}

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Select member *</span>
          <select
            autoFocus
            required
            value={assigneeId}
            onChange={(e) => setAssigneeId(e.target.value)}
            disabled={isLoading}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground disabled:opacity-50"
          >
            <option value="">— pick a member —</option>
            {members.map((m) => (
              <option key={m.id} value={m.id}>
                {m.display_name || m.email}
              </option>
            ))}
          </select>
        </label>

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
            disabled={assignTask.isPending || !assigneeId}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            {assignTask.isPending ? "Saving…" : "Assign"}
          </button>
        </div>
      </form>
    </div>
  );
}
