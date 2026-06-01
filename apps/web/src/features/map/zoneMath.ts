import type { Zone } from "./types";

/** Parse a #rrggbb hex into an rgba() string. Pure. */
export function rgba(hex: string, alpha: number): string {
  const h = hex.replace("#", "");
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/** Fill ratio (0..1) of a zone. Pure. */
export function fillRatio(z: Zone): number {
  if (!z.capacity || z.items == null) return 0;
  return Math.min(1, z.items / z.capacity);
}

/** Slot status from on-hand quantity. Pure. */
export function slotStatus(qty: number): "out" | "low" | "ok" {
  if (qty <= 0) return "out";
  if (qty < 10) return "low";
  return "ok";
}
