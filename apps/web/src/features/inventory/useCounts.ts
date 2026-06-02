import { useMutation } from "@tanstack/react-query";
import { useApi } from "../../lib/api";

/** One SKU line in a count session. system_qty is 0 while blind/open. */
export interface CountLine {
  sku: string;
  system_qty: number;
  counted_qty?: number;
}

/** A cycle-count session. */
export interface CountSession {
  id: string;
  zone_id: string;
  status: "open" | "completed";
  lines: CountLine[];
}

/** One reconciled line in the completion summary. */
export interface CountVariance {
  sku: string;
  system_qty: number;
  counted_qty: number;
  variance: number;
}

export interface CompleteCountResult {
  id: string;
  status: string;
  reconciled: number;
  variances: CountVariance[];
}

function countsPath(storeId: string): string {
  return `/api/v1/stores/${storeId}/counts`;
}

/** Start a blind cycle count for a zone (snapshots on-hand). */
export function useStartCount(storeId: string) {
  const { post } = useApi();
  return useMutation({
    mutationFn: (zoneId: string) => post<CountSession>(countsPath(storeId), { zone_id: zoneId }),
  });
}

/** Record the blind physical count for one SKU. */
export function useSetCountLine(storeId: string, countId: string) {
  const { patch } = useApi();
  return useMutation({
    mutationFn: (body: { sku: string; counted_qty: number }) =>
      patch(`${countsPath(storeId)}/${countId}/lines`, body),
  });
}

/** Complete a count: apply counted qty, return variances. */
export function useCompleteCount(storeId: string, countId: string) {
  const { post } = useApi();
  return useMutation({
    mutationFn: () => post<CompleteCountResult>(`${countsPath(storeId)}/${countId}/complete`, {}),
  });
}
