---
name: react-patterns
description: React + Tailwind + Zustand + TanStack Query patterns for live-rack. Use when building feature modules, shared components, or stores. References shell.jsx and screens-*.jsx for visual ground truth.
---

# React Patterns — live-rack

## Feature Module Structure

```
apps/web/src/features/map/
  index.ts              # public exports only
  MapPage.tsx           # page component (data fetching here)
  ZoneCanvas.tsx        # Konva canvas
  ZoneCard.tsx          # pure presentational
  ZoneDetailSidebar.tsx
  useZones.ts           # TanStack Query hook
  useZoneEditor.ts      # local edit state (Zustand slice)
  map.store.ts          # Zustand slice
  __tests__/
    ZoneCard.test.tsx
    useZones.test.ts
```

## Data Fetching (TanStack Query)

```typescript
// features/map/useZones.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'

export const zoneKeys = {
  all: ['zones'] as const,
  byStore: (storeId: string) => [...zoneKeys.all, storeId] as const,
}

export function useZones(storeId: string) {
  return useQuery({
    queryKey: zoneKeys.byStore(storeId),
    queryFn: () => api.get<Zone[]>(`/zones?storeId=${storeId}`),
  })
}

export function useUpsertZone() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (zone: Zone) => api.put<Zone>(`/zones/${zone.id}`, zone),
    onSuccess: (_, zone) => qc.invalidateQueries({ queryKey: zoneKeys.byStore(zone.storeId) }),
  })
}
```

## Zustand Store Slice

```typescript
// features/map/map.store.ts
interface MapState {
  selectedZoneId: string | null
  editingZone: Zone | null
  setSelectedZone: (id: string | null) => void
  startEdit: (zone: Zone) => void
  cancelEdit: () => void
}

export const useMapStore = create<MapState>((set) => ({
  selectedZoneId: null,
  editingZone: null,
  setSelectedZone: (id) => set({ selectedZoneId: id }),
  startEdit: (zone) => set({ editingZone: zone }),
  cancelEdit: () => set({ editingZone: null }),
}))
```

## Page Component (all 4 states)

```typescript
// features/map/MapPage.tsx
export function MapPage() {
  const { storeId } = useAuth()
  const { data: zones, isLoading, isError, error } = useZones(storeId)

  if (isLoading) return <ZoneCanvasSkeleton />
  if (isError) return <ErrorBanner message={error.message} />
  if (!zones?.length) return <EmptyZones />

  return <ZoneCanvas zones={zones} />
}
```

## Tailwind Token Usage

Use CSS vars from `tokens.css` — never hardcode colors:
```tsx
// Correct
<div className="bg-[--accent] text-[--text-primary]" />

// Wrong — hardcoded
<div style={{ background: '#2563eb' }} />
```

## WebSocket Real-Time Updates

```typescript
// lib/ws.ts
export function useZoneUpdates(orgId: string) {
  const qc = useQueryClient()
  useEffect(() => {
    const ws = new WebSocket(`/ws?org=${orgId}`)
    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data)
      if (msg.type === 'zone.updated') {
        qc.setQueryData(zoneKeys.byStore(msg.storeId), (old: Zone[]) =>
          old?.map(z => z.id === msg.zone.id ? msg.zone : z) ?? [msg.zone]
        )
      }
    }
    return () => ws.close()
  }, [orgId, qc])
}
```
