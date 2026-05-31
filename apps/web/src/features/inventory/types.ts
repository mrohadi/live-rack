/** Item lifecycle state — mirrors the items.status CHECK constraint. */
export type ItemStatus = "active" | "discontinued" | "recalled";

/** Sales-velocity band. Computed from rolling 7d sales in LR-304; absent until then. */
export type VelocityBand = "hot" | "warm" | "cold" | "dead";

export interface InventoryRow {
  id: string;
  zone_id: string;
  sku: string;
  name: string;
  category: string;
  status: string;
  qty: number;
  updated_at: string;
  /** Populated by LR-304; treated as "cold" while undefined. */
  velocity?: VelocityBand;
}
