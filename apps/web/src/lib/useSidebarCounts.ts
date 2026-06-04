import { useQuery } from "@tanstack/react-query";
import { useApi } from "./api";
import { useCurrentStoreData } from "../features/stores/useStores";

/** Stale time 5 min — sidebar badges do not need live accuracy. */
const STALE = 5 * 60_000;

function fmt(n: number | undefined): string | null {
  if (n === undefined) return null;
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return String(n);
}

export function useSidebarCounts() {
  const { get } = useApi();
  const store = useCurrentStoreData();
  const sid = store?.id ?? "";
  const enabled = Boolean(sid);

  const zones = useQuery({
    queryKey: ["sidebar", "zones", sid],
    queryFn: () => get<unknown[]>(`/api/v1/${sid}/zones`),
    enabled,
    staleTime: STALE,
    select: (d) => d.length,
  });

  const inventory = useQuery({
    queryKey: ["sidebar", "inventory", sid],
    queryFn: () => get<unknown[]>(`/api/v1/${sid}/inventory`),
    enabled,
    staleTime: STALE,
    select: (d) => d.length,
  });

  const tasks = useQuery({
    queryKey: ["sidebar", "tasks", sid],
    queryFn: () => get<unknown[]>(`/api/v1/${sid}/tasks`),
    enabled,
    staleTime: STALE,
    select: (d) => d.length,
  });

  const integrations = useQuery({
    queryKey: ["sidebar", "integrations"],
    queryFn: () => get<unknown[]>("/api/v1/integrations"),
    staleTime: STALE,
    select: (d) => d.length,
  });

  return {
    zones: fmt(zones.data),
    inventory: fmt(inventory.data),
    tasks: fmt(tasks.data),
    integrations: fmt(integrations.data),
  };
}
