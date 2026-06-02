/** Lifecycle of a pick list. */
export type PickListStatus = "open" | "picking" | "completed" | "cancelled";

/** Lifecycle of one pick line. */
export type PickLineStatus = "pending" | "picked" | "short";

/** One stop on the optimised pick route. */
export interface PickLine {
  id: string;
  seq: number;
  sku: string;
  zone_id?: string;
  zone_name: string;
  zone_x: number;
  zone_y: number;
  qty_requested: number;
  qty_picked: number;
  status: PickLineStatus;
}

/** A pick list with its ordered route. */
export interface PickBoard {
  id: string;
  reference: string;
  status: PickListStatus;
  lines: PickLine[];
}

/** One pick list in the index. */
export interface PickListSummary {
  id: string;
  reference: string;
  status: PickListStatus;
  line_count: number;
  done_count: number;
  created_at: string;
}

/** A requested line when creating a pick list. */
export interface NewPickLine {
  sku: string;
  qty: number;
}
