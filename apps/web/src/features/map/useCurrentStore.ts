// Placeholder until org→store selection lands. The Vite env supplies a
// dev/E2E store id; everything store-scoped funnels through this hook so
// it's a one-line swap later.
export function useCurrentStore(): string {
  return import.meta.env.VITE_STORE_ID ?? "00000000-0000-0000-0000-000000000001";
}
