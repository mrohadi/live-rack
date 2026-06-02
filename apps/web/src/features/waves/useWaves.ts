import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { PickLine } from "../picking/types";
import type { WaveBoard, WaveStop, WaveSummary } from "./types";

/** Query-key factory. */
export const waveKeys = {
  all: ["waves"] as const,
  list: (storeId: string) => [...waveKeys.all, "list", storeId] as const,
  board: (storeId: string, id: string) => [...waveKeys.all, "board", storeId, id] as const,
};

function basePath(storeId: string): string {
  return `/api/v1/stores/${storeId}/waves`;
}

/** Wave progress over merged stops. Pure. */
export function waveProgress(stops: WaveStop[]): { done: number; total: number; pct: number } {
  const total = stops.length;
  const done = stops.reduce((n, s) => (s.status === "pending" ? n : n + 1), 0);
  const pct = total === 0 ? 0 : Math.round((done / total) * 100);
  return { done, total, pct };
}

/** Adapt a merged stop to the PickLine shape the route map renders. Pure. */
export function stopToLine(stop: WaveStop): PickLine {
  return {
    id: `${stop.sku}|${stop.zone_id ?? ""}`,
    seq: stop.seq,
    sku: stop.sku,
    zone_id: stop.zone_id,
    zone_name: stop.zone_name,
    zone_x: stop.zone_x,
    zone_y: stop.zone_y,
    qty_requested: stop.qty_requested,
    qty_picked: stop.qty_picked,
    status: stop.status,
  };
}

/** Fetch the wave index for a store. */
export function useWaves(storeId: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: waveKeys.list(storeId),
    queryFn: () => get<WaveSummary[]>(basePath(storeId)),
  });
}

/** Fetch a single wave with its merged route. */
export function useWaveBoard(storeId: string, id: string | undefined) {
  const { get } = useApi();
  return useQuery({
    queryKey: waveKeys.board(storeId, id ?? ""),
    queryFn: () => get<WaveBoard>(`${basePath(storeId)}/${id}`),
    enabled: Boolean(id),
  });
}

/** Create a wave from pick lists, then refresh the index. */
export function useCreateWave(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { reference: string; list_ids: string[] }) =>
      post<WaveBoard>(basePath(storeId), body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: waveKeys.list(storeId) }),
  });
}

/** Mark a wave as in-progress. */
export function useStartWave(storeId: string, id: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => post(`${basePath(storeId)}/${id}/start`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: waveKeys.board(storeId, id) });
      void qc.invalidateQueries({ queryKey: waveKeys.list(storeId) });
    },
  });
}

/** Confirm a merged stop; server allocates the qty across member orders. */
export function useConfirmStop(storeId: string, id: string) {
  const { patch } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { sku: string; zone_id: string; qty_picked: number }) =>
      patch<WaveBoard>(`${basePath(storeId)}/${id}/stops`, body),
    onSuccess: (board) => qc.setQueryData(waveKeys.board(storeId, id), board),
  });
}

/** Close a wave, then refresh index + board. */
export function useCompleteWave(storeId: string, id: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => post(`${basePath(storeId)}/${id}/complete`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: waveKeys.board(storeId, id) });
      void qc.invalidateQueries({ queryKey: waveKeys.list(storeId) });
    },
  });
}
