import type { PickLineStatus, PickListStatus } from "../picking/types";

/** A merged stop on the wave route (summed demand for one SKU+zone). */
export interface WaveStop {
  seq: number;
  sku: string;
  zone_id?: string;
  zone_name: string;
  zone_x: number;
  zone_y: number;
  qty_requested: number;
  qty_picked: number;
  order_count: number;
  status: PickLineStatus;
}

/** A wave with its merged, route-ordered stops. */
export interface WaveBoard {
  id: string;
  reference: string;
  status: PickListStatus;
  stops: WaveStop[];
}

/** One wave in the index. */
export interface WaveSummary {
  id: string;
  reference: string;
  status: PickListStatus;
  list_count: number;
  created_at: string;
}
