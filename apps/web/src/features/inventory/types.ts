/** Item lifecycle state — mirrors the items.status CHECK constraint. */
export type ItemStatus = "active" | "discontinued" | "recalled";

/** Sales-velocity band. Computed from rolling 7d sales in LR-304; absent until then. */
export type VelocityBand = "hot" | "warm" | "cold" | "dead";

/** On-hand stock band relative to the SKU's reorder point (LR-305). */
export type StockStatus = "in_stock" | "low" | "out";

export interface InventoryRow {
  id: string;
  zone_id: string;
  sku: string;
  name: string;
  category: string;
  status: string;
  qty: number;
  /** Qty at/below which restock triggers; 0 disables (LR-305). */
  reorder_point?: number;
  /** Derived band from qty vs reorder_point (LR-305). */
  stock_status?: StockStatus;
  updated_at: string;
  /** Populated by LR-304; treated as "cold" while undefined. */
  velocity?: VelocityBand;
}
