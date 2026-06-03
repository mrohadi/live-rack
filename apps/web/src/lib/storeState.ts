const KEY = "lr:store_id";

export function getSelectedStoreId(): string | null {
  return localStorage.getItem(KEY);
}

export function setSelectedStoreId(id: string): void {
  localStorage.setItem(KEY, id);
}
