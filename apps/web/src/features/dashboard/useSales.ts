import { useQuery } from "@tanstack/react-query";
import { useApi } from "../../lib/api";

export interface SalesSummary {
  revenue_cents: number;
  units: number;
  orders: number;
  spark: number[];
}

export const salesKeys = {
  summary: ["sales", "summary"] as const,
};

/** Format minor units (cents) as a localised currency string. Pure. */
export function formatCents(cents: number, currency = "USD"): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency }).format(cents / 100);
}

/** Build an SVG polyline points string for a sparkline. Pure. */
export function sparkPoints(values: number[], width: number, height: number): string {
  if (values.length === 0) return "";
  const max = Math.max(...values, 1);
  const step = values.length > 1 ? width / (values.length - 1) : 0;
  return values
    .map((v, i) => `${(i * step).toFixed(1)},${(height - (v / max) * height).toFixed(1)}`)
    .join(" ");
}

/** Fetch the org's sales summary widgets. */
export function useSalesSummary() {
  const { get } = useApi();
  return useQuery({
    queryKey: salesKeys.summary,
    queryFn: () => get<SalesSummary>("/api/v1/sales/summary"),
  });
}
