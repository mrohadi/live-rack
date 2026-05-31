import type { ScanPayload } from "@/lib/scanQueue";
import { enqueueScan, flushQueue, pendingCount } from "@/lib/scanQueue";
import { useCallback, useEffect, useState } from "react";
import { useApi } from "../../lib/api";
import { useBarcodeScanner } from "./useBarcodeScanner";
import { useZebraHID } from "./useZebraHID";

export function ScannerPage() {
  const [active, setActive] = useState(false);
  const [scans, setScans] = useState<string[]>([]);
  const [pending, setPending] = useState(0);
  const api = useApi();

  const handleScan = useCallback(
    async (sku: string) => {
      setScans((prev) => (prev[0] === sku ? prev : [sku, ...prev].slice(0, 20)));
      const payload: ScanPayload = { sku, zoneId: "", scannedAt: Date.now() };

      try {
        await api.post("/scan", payload);
      } catch {
        await enqueueScan(payload);
        setPending(await pendingCount());
      }
    },
    [api],
  );

  useEffect(() => {
    const sync = async () => {
      try {
        await flushQueue((s) => api.post("/scan", s));
      } finally {
        setPending(await pendingCount());
      }
    };
    void sync();
    window.addEventListener("online", sync);
    return () => window.removeEventListener("online", sync);
  }, [api]);

  const videoRef = useBarcodeScanner({ onScan: handleScan, active });
  const { connect } = useZebraHID({ onScan: handleScan });

  {
    pending > 0 && (
      <span className="rounded bg-amber-100 px-2 py-1 text-xs text-amber-800">
        {pending} queued offline
      </span>
    );
  }

  return (
    <div>
      <div className="page-head">
        <div>
          <div className="page-title">Scanner</div>
          <div className="page-sub">Camera + WebHID barcode scanning — P2</div>
        </div>
        <div className="flex gap-2">
          <button onClick={() => setActive((a) => !a)}>
            {active ? "Stop camera" : "Start camera"}
          </button>
          <button onClick={connect}>Connect Zebra</button>
        </div>
      </div>

      <video
        ref={videoRef}
        className="w-full max-w-md rounded-lg bg-black aspect-video"
        muted
        playsInline
      />

      <ul className="mt-4 space-y-1">
        {scans.map((sku, i) => (
          <li key={`${sku}-${i}`} className="rounded bg-slate-100 px-3 py-2 font-mono text-sm">
            {sku}
          </li>
        ))}
      </ul>
    </div>
  );
}
