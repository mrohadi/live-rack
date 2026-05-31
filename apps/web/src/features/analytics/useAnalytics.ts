import { useQuery } from "@tanstack/react-query";
import { useApi } from "../../lib/api";

/** 7x24 scan heatmap: grid[dayIndex][hour], dayIndex 0 = Monday. */
export interface HeatmapResponse {
  grid: number[][];
  max: number;
}

export const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"] as const;
export const HOURS_IN_DAY = 24;

export const analyticsKeys = {
  heatmap: (zoneId?: string) => ["analytics", "heatmap", zoneId ?? "all"] as const,
};

/** Normalise a cell value to 0..1 of the peak. Pure. */
export function cellIntensity(value: number, max: number): number {
  if (max <= 0) return 0;
  const r = value / max;
  return r < 0 ? 0 : r > 1 ? 1 : r;
}

/** Background for a heatmap cell, blending the accent over the panel by
 *  intensity. Pure. */
export function heatColor(value: number, max: number, accent: string): string {
  const pct = Math.round(cellIntensity(value, max) * 100);
  return `color-mix(in oklab, ${accent} ${pct}%, var(--panel))`;
}

/** Fetch the org's 7x24 scan heatmap, optionally scoped to one zone. */
export function useHeatmap(zoneId?: string) {
  const { get } = useApi();
  return useQuery({
    queryKey: analyticsKeys.heatmap(zoneId),
    queryFn: () =>
      get<HeatmapResponse>(`/api/v1/analytics/heatmap${zoneId ? `?zone_id=${zoneId}` : ""}`),
  });
}
