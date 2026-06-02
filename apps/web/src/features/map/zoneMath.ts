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

/** Zone fill palette (matches the legend). */
export const ZONE_COLORS = [
  "#2563eb",
  "#0891b2",
  "#16a34a",
  "#7c3aed",
  "#d97706",
  "#dc2626",
  "#5b6577",
];

/** Pick a random zone color. */
export function randomZoneColor(): string {
  return ZONE_COLORS[Math.floor(Math.random() * ZONE_COLORS.length)];
}

interface Rect {
  x: number;
  y: number;
  width: number;
  height: number;
}

/** True when two rects overlap (with an optional gap). Pure. */
export function rectsOverlap(a: Rect, b: Rect, gap = 0): boolean {
  return (
    a.x < b.x + b.width + gap &&
    a.x + a.width + gap > b.x &&
    a.y < b.y + b.height + gap &&
    a.y + a.height + gap > b.y
  );
}

/** Scan the canvas (0..100) for the first open slot that fits w×h without
 *  overlapping existing zones; falls back to the origin. Pure. */
export function findOpenSlot(
  existing: Rect[],
  w: number,
  h: number,
  step = 2,
  gap = 2,
): { x: number; y: number } {
  for (let y = 0; y <= 100 - h; y += step) {
    for (let x = 0; x <= 100 - w; x += step) {
      const candidate = { x, y, width: w, height: h };
      if (!existing.some((z) => rectsOverlap(candidate, z, gap))) return { x, y };
    }
  }
  return { x: 0, y: 0 };
}
