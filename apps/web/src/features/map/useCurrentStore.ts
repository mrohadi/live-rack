import { useCurrentStoreData } from "../stores/useStores";

/** Returns the currently selected store's UUID, or empty string while loading. */
export function useCurrentStore(): string {
  const store = useCurrentStoreData();
  return store?.id ?? "";
}
