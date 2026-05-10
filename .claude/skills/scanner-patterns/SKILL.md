---
name: scanner-patterns
description: Scanner PWA patterns — camera barcode via @zxing/browser, WebHID for Zebra USB scanners, IndexedDB offline queue with sync-on-reconnect. Use when building scanner UI, validation hooks, or offline queue logic.
---

# Scanner Patterns — live-rack

## Camera Barcode (@zxing/browser)

```typescript
// features/scanner/useBarcodeScanner.ts
import { BrowserMultiFormatReader } from '@zxing/browser'

export function useBarcodeScanner(onScan: (sku: string) => void) {
  const readerRef = useRef<BrowserMultiFormatReader>()
  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    const reader = new BrowserMultiFormatReader()
    readerRef.current = reader

    reader.decodeFromVideoDevice(undefined, videoRef.current!, (result, err) => {
      if (result) onScan(result.getText())
    })

    return () => reader.reset()
  }, [onScan])

  return videoRef
}
```

## WebHID Zebra Scanner

```typescript
// features/scanner/useZebraHID.ts
export function useZebraHID(onScan: (sku: string) => void) {
  useEffect(() => {
    let device: HIDDevice | null = null
    let buffer = ''

    async function connect() {
      const devices = await navigator.hid.requestDevice({ filters: [{ vendorId: 0x05E0 }] })
      device = devices[0]
      if (!device) return
      await device.open()
      device.addEventListener('inputreport', (e: HIDInputReportEvent) => {
        const char = String.fromCharCode(new DataView(e.data.buffer).getUint8(2))
        if (char === '\r') { onScan(buffer.trim()); buffer = '' }
        else buffer += char
      })
    }

    connect().catch(console.error)
    return () => { device?.close() }
  }, [onScan])
}
```

## Offline IndexedDB Queue

```typescript
// lib/scanQueue.ts
import { openDB } from 'idb'

const DB_NAME = 'live-rack-scanner'
const STORE = 'pending-scans'

export async function enqueueScan(scan: ScanPayload) {
  const db = await openDB(DB_NAME, 1, {
    upgrade(db) { db.createObjectStore(STORE, { keyPath: 'id', autoIncrement: true }) },
  })
  await db.add(STORE, { ...scan, queuedAt: Date.now() })
}

export async function flushQueue(api: ApiClient) {
  const db = await openDB(DB_NAME, 1)
  const all = await db.getAll(STORE)
  for (const scan of all) {
    try {
      await api.post('/scan', scan)
      await db.delete(STORE, scan.id)
    } catch { break } // stop on first failure, retry next reconnect
  }
}
```

## Validation Rules Engine

```typescript
// features/scanner/validation.ts
type ValidationRule = (scan: ScanAttempt, zone: Zone, item: Item) => ValidationResult

export const rules: ValidationRule[] = [
  categoryMatch,    // item.type must be allowed by zone.allowedTypes
  capacityCheck,   // zone.items + 1 must not exceed zone.capacity
  dwellCheck,      // item must not have been scanned into this zone in last 60s (anti-double-scan)
  zoneActiveCheck, // zone must not be "closed" (outbound sealed)
]

export function validate(scan: ScanAttempt, zone: Zone, item: Item): ValidationResult {
  for (const rule of rules) {
    const result = rule(scan, zone, item)
    if (!result.valid) return result
  }
  return { valid: true }
}
```

## Sync on Reconnect

```typescript
// features/scanner/ScannerPage.tsx
useEffect(() => {
  const handleOnline = async () => {
    await flushQueue(api)
    toast.success('Offline scans synced')
  }
  window.addEventListener('online', handleOnline)
  return () => window.removeEventListener('online', handleOnline)
}, [api])
```
