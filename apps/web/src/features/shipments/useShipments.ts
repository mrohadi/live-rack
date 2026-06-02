import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { ShipmentBoard, ShipmentStatus, ShipmentSummary } from "./types";

/** Query-key factory. */
export const shipmentKeys = {
  all: ["shipments"] as const,
  list: (storeId: string) => [...shipmentKeys.all, "list", storeId] as const,
  board: (storeId: string, id: string) => [...shipmentKeys.all, "board", storeId, id] as const,
};

function basePath(storeId: string): string {
  return `/api/v1/stores/${storeId}/shipments`;
}

/** Human label for a shipment status. Pure. */
export function shipmentStatusLabel(status: ShipmentStatus): string {
  switch (status) {
    case "packing":
      return "Packing";
    case "packed":
      return "Packed";
    case "dispatched":
      return "Dispatched";
    default:
      return "Cancelled";
  }
}

/** Fetch the shipment index for a store. */
export function useShipments(storeId: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: shipmentKeys.list(storeId),
    queryFn: () => get<ShipmentSummary[]>(basePath(storeId)),
  });
}

/** Fetch a single shipment with its items. */
export function useShipmentBoard(storeId: string, id: string | undefined) {
  const { get } = useApi();
  return useQuery({
    queryKey: shipmentKeys.board(storeId, id ?? ""),
    queryFn: () => get<ShipmentBoard>(`${basePath(storeId)}/${id}`),
    enabled: Boolean(id),
  });
}

/** Create a shipment from a completed pick list, then refresh the index. */
export function useCreateShipment(storeId: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { pick_list_id: string; reference: string }) =>
      post<ShipmentBoard>(basePath(storeId), body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: shipmentKeys.list(storeId) }),
  });
}

function useShipmentAction(storeId: string, id: string, path: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body?: unknown) => post(`${basePath(storeId)}/${id}/${path}`, body ?? {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: shipmentKeys.board(storeId, id) });
      void qc.invalidateQueries({ queryKey: shipmentKeys.list(storeId) });
    },
  });
}

/** Mark a shipment packed. */
export function usePackShipment(storeId: string, id: string) {
  return useShipmentAction(storeId, id, "pack");
}

/** Dispatch a packed shipment with carrier + tracking. */
export function useDispatchShipment(storeId: string, id: string) {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { carrier: string; tracking_number: string }) =>
      post<ShipmentBoard>(`${basePath(storeId)}/${id}/dispatch`, body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: shipmentKeys.board(storeId, id) });
      void qc.invalidateQueries({ queryKey: shipmentKeys.list(storeId) });
    },
  });
}

/** Cancel a shipment before it ships. */
export function useCancelShipment(storeId: string, id: string) {
  return useShipmentAction(storeId, id, "cancel");
}
