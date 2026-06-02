import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import { TASK_COLUMNS, type Task, type TaskPriority, type TaskStatus } from "./types";

/** Query-key factory — keeps cache keys consistent across hooks. */
export const taskKeys = {
  all: ["tasks"] as const,
  list: (storeId: string) => [...taskKeys.all, "list", storeId] as const,
};

function tasksPath(storeId: string): string {
  return `/api/v1/stores/${storeId}/tasks`;
}

/** Group tasks into kanban columns, preserving column order. Pure. */
export function groupByStatus(tasks: Task[]): Record<TaskStatus, Task[]> {
  const out = {} as Record<TaskStatus, Task[]>;
  for (const { status } of TASK_COLUMNS) out[status] = [];
  for (const t of tasks) out[t.status]?.push(t);
  return out;
}

/** Return next state with the given task moved to a new column. Pure — no-op if unchanged or missing. */
export function moveTask(tasks: Task[], id: string, status: TaskStatus): Task[] {
  return tasks.map((t) => (t.id === id && t.status !== status ? { ...t, status } : t));
}

export interface CreateTaskInput {
  zone_id?: string;
  title: string;
  priority: TaskPriority;
  due_at?: string;
}

/** Create a new task, optionally scoped to a zone. */
export function useCreateTask(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateTaskInput) => post<Task>(tasksPath(storeId), input),
    onSettled: () => void qc.invalidateQueries({ queryKey: taskKeys.list(storeId) }),
  });
}

/** Fetch the task board for a store. */
export function useTasks(storeId: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: taskKeys.list(storeId),
    queryFn: () => get<Task[]>(tasksPath(storeId)),
  });
}

/** Move a task to a new column, optimistically updating the cached board. */
export function useMoveTask(storeId: string) {
  const { patch } = useApi();
  const qc = useQueryClient();
  const key = taskKeys.list(storeId);

  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: TaskStatus }) =>
      patch<Task>(`${tasksPath(storeId)}/${id}`, { status }),
    onMutate: async ({ id, status }) => {
      await qc.cancelQueries({ queryKey: key });
      const prev = qc.getQueryData<Task[]>(key);
      qc.setQueryData<Task[]>(key, (cur) => (cur ? moveTask(cur, id, status) : cur));
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(key, ctx.prev);
    },
    onSettled: () => void qc.invalidateQueries({ queryKey: key }),
  });
}
