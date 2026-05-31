/** Entity kind a search hit points at. Mirrors the API `kind` column. */
export type SearchKind = "item" | "zone";

/** One ⌘K search hit returned by GET /api/v1/search. */
export interface SearchResult {
  kind: SearchKind;
  id: string;
  label: string;
  sublabel: string;
  score: number;
}
