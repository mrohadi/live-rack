import { openDB, type DBSchema, type IDBPDatabase } from "idb";

let dbPromise: Promise<IDBPDatabase<ScanQueueDB>> | null = null;

function db(): Promise<IDBPDatabase<ScanQueueDB>> {
  if (!dbPromise) {
    dbPromise = openDB<ScanQueueDB>(DB_NAME, 1, {
      upgrade(d) {
        d.createObjectStore(STORE, { keyPath: "id", autoIncrement: true });
      },
    });
  }
  return dbPromise;
}

export async function closeQueue(): Promise<void> {
  if (dbPromise) {
    (await dbPromise).close();
    dbPromise = null;
  }
}

export interface ScanPayload {
  sku: string;
  zoneId: string;
  scannedAt: number;
}

interface ScanQueueDB extends DBSchema {
  "pending-scans": {
    key: number;
    value: ScanPayload & { queuedAt: number };
  };
}

const DB_NAME = "live-rack-scanner";
const STORE = "pending-scans";

export async function enqueueScan(scan: ScanPayload): Promise<void> {
  await (await db()).add(STORE, { ...scan, queuedAt: Date.now() });
}

export async function pendingCount(): Promise<number> {
  return (await db()).count(STORE);
}

export async function flushQueue(send: (s: ScanPayload) => Promise<void>): Promise<void> {
  const d = await db();
  const keys = await d.getAllKeys(STORE);
  for (const key of keys) {
    const scan = await d.get(STORE, key);
    if (!scan) continue;
    await send(scan);
    await d.delete(STORE, key);
  }
}
