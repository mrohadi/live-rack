export type ShipmentStatus = "packing" | "packed" | "dispatched" | "cancelled";

/** One packed line on a shipment. */
export interface ShipmentItem {
  sku: string;
  qty: number;
}

/** A shipment with its packed items. */
export interface ShipmentBoard {
  id: string;
  reference: string;
  status: ShipmentStatus;
  carrier: string;
  tracking_number: string;
  items: ShipmentItem[];
}

/** One shipment in the index. */
export interface ShipmentSummary {
  id: string;
  reference: string;
  status: ShipmentStatus;
  carrier: string;
  tracking_number: string;
  item_count: number;
  created_at: string;
}
