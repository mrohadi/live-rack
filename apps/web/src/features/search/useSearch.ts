import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useApi } from "../../lib/api";
import type { SearchResult } from "./types";

/** Below this length we skip the request — matches the API's 2-char minimum. */
export const MIN_QUERY_LENGTH = 2;

/** Query-key factory — keeps cache keys consistent across hooks. */
export const searchKeys = {
  all: ["search"] as const,
  query: (q: string) => [...searchKeys.all, q] as const,
};

/** Build the search request path. Exported for tests. */
export function searchPath(query: string, limit = 20): string {
  const params = new URLSearchParams({ q: query, limit: String(limit) });
  return `/api/v1/search?${params.toString()}`;
}

/** Debounce a value by `delay` ms. */
export function useDebounced<T>(value: T, delay = 150): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(t);
  }, [value, delay]);
  return debounced;
}

/** Fuzzy ⌘K search over items and zones. Debounced; idle until 2+ chars. */
export function useSearch(rawQuery: string) {
  const { get } = useApi();
  const query = useDebounced(rawQuery.trim());
  const enabled = query.length >= MIN_QUERY_LENGTH;

  return useQuery({
    queryKey: searchKeys.query(query),
    queryFn: () => get<SearchResult[]>(searchPath(query)),
    enabled,
    staleTime: 10_000,
  });
}
