import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { Board, Card } from "./types";

/** Query-key factory — keeps cache keys consistent across hooks. */
export const pipelineKeys = {
  all: ["pipelines"] as const,
  list: (storeId: string) => [...pipelineKeys.all, "list", storeId] as const,
  board: (storeId: string, pipelineId: string) =>
    [...pipelineKeys.all, "board", storeId, pipelineId] as const,
};

function basePath(storeId: string): string {
  return `/api/v1/stores/${storeId}/pipelines`;
}

/** Bucket cards by their stage position. Pure. Every stage gets an entry. */
export function cardsByStage(
  stagePositions: number[],
  cards: Card[],
): Record<number, Card[]> {
  const out: Record<number, Card[]> = {};
  for (const pos of stagePositions) out[pos] = [];
  for (const c of cards) (out[c.stage_position] ??= []).push(c);
  return out;
}

/** Return next cards with the given card moved to a new stage. Pure — resets ageing. */
export function moveCard(cards: Card[], id: string, stagePosition: number): Card[] {
  return cards.map((c) =>
    c.id === id && c.stage_position !== stagePosition
      ? { ...c, stage_position: stagePosition, age_seconds: 0, ageing: false }
      : c,
  );
}

/** Count cards breaching their stage SLA. Pure. */
export function ageingCount(cards: Card[]): number {
  return cards.reduce((n, c) => (c.ageing ? n + 1 : n), 0);
}

/** Human-readable dwell time, e.g. "3h", "2d", "45m". Pure. */
export function formatAge(seconds: number): string {
  if (seconds >= 86400) return `${Math.floor(seconds / 86400)}d`;
  if (seconds >= 3600) return `${Math.floor(seconds / 3600)}h`;
  if (seconds >= 60) return `${Math.floor(seconds / 60)}m`;
  return `${seconds}s`;
}

/** Instantiate a pipeline from a built-in template, then refresh the list. */
export function useCreateFromTemplate(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (templateKey: string) =>
      post<Board["pipeline"]>(`${basePath(storeId)}/from-template`, {
        template_key: templateKey,
      }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: pipelineKeys.list(storeId) }),
  });
}

/** Fetch the pipeline selector list for a store. */
export function usePipelines(storeId: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: pipelineKeys.list(storeId),
    queryFn: () => get<Pick<Board["pipeline"], "id" | "key" | "name">[]>(basePath(storeId)),
  });
}

/** Fetch a single pipeline board (stages + cards). */
export function useBoard(storeId: string, pipelineId: string | undefined) {
  const { get } = useApi();
  return useQuery({
    queryKey: pipelineKeys.board(storeId, pipelineId ?? ""),
    queryFn: () => get<Board>(`${basePath(storeId)}/${pipelineId}`),
    enabled: Boolean(pipelineId),
  });
}

/** Move a card to a new stage, optimistically updating the cached board. */
export function useMoveCard(storeId: string, pipelineId: string) {
  const { patch } = useApi();
  const qc = useQueryClient();
  const key = pipelineKeys.board(storeId, pipelineId);

  return useMutation({
    mutationFn: ({ id, stagePosition }: { id: string; stagePosition: number }) =>
      patch<Card>(`${basePath(storeId)}/${pipelineId}/cards/${id}`, {
        stage_position: stagePosition,
      }),
    onMutate: async ({ id, stagePosition }) => {
      await qc.cancelQueries({ queryKey: key });
      const prev = qc.getQueryData<Board>(key);
      qc.setQueryData<Board>(key, (cur) =>
        cur ? { ...cur, cards: moveCard(cur.cards, id, stagePosition) } : cur,
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(key, ctx.prev);
    },
    onSettled: () => void qc.invalidateQueries({ queryKey: key }),
  });
}
