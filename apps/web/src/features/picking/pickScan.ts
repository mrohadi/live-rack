import type { PickLine } from "./types";

/** Normalise a scanned code for comparison (trim, upper-case). Pure. */
export function normalizeScan(code: string): string {
  return code.trim().toUpperCase();
}

/**
 * Resolve a scanned SKU to the line it should credit: the lowest-seq line still
 * pending for that SKU. Returns undefined on a mis-scan (no pending line for the
 * code). Pure.
 */
export function resolveScan(code: string, lines: PickLine[]): PickLine | undefined {
  const want = normalizeScan(code);
  return [...lines]
    .sort((a, b) => a.seq - b.seq)
    .find((l) => l.status === "pending" && normalizeScan(l.sku) === want);
}

/** Next cumulative picked qty after one scan, capped at the requested qty. Pure. */
export function nextScanQty(current: number, requested: number): number {
  return Math.min(current + 1, requested);
}
