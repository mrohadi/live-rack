import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { NewPickLine, PickBoard, PickLine, PickListSummary } from "./types";

/** Query-key factory — keeps cache keys consistent across hooks. */
export const pickKeys = {
  all: ["picking"] as const,
  list: (storeId: string) => [...pickKeys.all, "list", storeId] as const,
  board: (storeId: string, id: string) => [...pickKeys.all, "board", storeId, id] as const,
};

function basePath(storeId: string): string {
  return `/api/v1/stores/${storeId}/pick-lists`;
}

/** Pick progress for a route. Pure. */
export function pickProgress(lines: PickLine[]): { done: number; total: number; pct: number } {
  const total = lines.length;
  const done = lines.reduce((n, l) => (l.status === "pending" ? n : n + 1), 0);
  const pct = total === 0 ? 0 : Math.round((done / total) * 100);
  return { done, total, pct };
}

/** First not-yet-picked stop, by route order. Pure. Undefined when route done. */
export function nextPendingLine(lines: PickLine[]): PickLine | undefined {
  return [...lines].sort((a, b) => a.seq - b.seq).find((l) => l.status === "pending");
}

/** Human label for a line status. Pure. */
export function lineStatusLabel(status: PickLine["status"]): string {
  switch (status) {
    case "picked":
      return "Picked";
    case "short":
      return "Short";
    default:
      return "Pending";
  }
}

/** Fetch the pick-list index for a store. */
export function usePickLists(storeId: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: pickKeys.list(storeId),
    queryFn: () => get<PickListSummary[]>(basePath(storeId)),
  });
}

/** Fetch a single pick list with its route. */
export function usePickBoard(storeId: string, id: string | undefined) {
  const { get } = useApi();
  return useQuery({
    queryKey: pickKeys.board(storeId, id ?? ""),
    queryFn: () => get<PickBoard>(`${basePath(storeId)}/${id}`),
    enabled: Boolean(id),
  });
}

/** Create a pick list (server optimises the route), then refresh the index. */
export function useCreatePickList(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { reference: string; lines: NewPickLine[] }) =>
      post<PickBoard>(basePath(storeId), body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: pickKeys.list(storeId) }),
  });
}

/** Mark a pick list as in-progress. */
export function useStartPick(storeId: string, id: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => post(`${basePath(storeId)}/${id}/start`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: pickKeys.board(storeId, id) });
      void qc.invalidateQueries({ queryKey: pickKeys.list(storeId) });
    },
  });
}

/** Confirm a pick for one line; refresh the board afterwards. */
export function useConfirmPick(storeId: string, id: string) {
  const { patch } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ lineId, qtyPicked }: { lineId: string; qtyPicked: number }) =>
      patch<PickLine>(`${basePath(storeId)}/${id}/lines/${lineId}`, { qty_picked: qtyPicked }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: pickKeys.board(storeId, id) }),
  });
}

/** Close a pick list, then refresh the index + board. */
export function useCompletePick(storeId: string, id: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => post(`${basePath(storeId)}/${id}/complete`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: pickKeys.board(storeId, id) });
      void qc.invalidateQueries({ queryKey: pickKeys.list(storeId) });
    },
  });
}
