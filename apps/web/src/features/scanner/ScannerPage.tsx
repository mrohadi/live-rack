import type { ScanPayload } from "@/lib/scanQueue";
import { enqueueScan, flushQueue, pendingCount } from "@/lib/scanQueue";
import { type FormEvent, useCallback, useEffect, useState } from "react";
import { useApi } from "../../lib/api";
import { useBarcodeScanner } from "./useBarcodeScanner";
import { useZebraHID } from "./useZebraHID";

// ScanDecision mirrors the API ValidateResponse — the server's mis-scan verdict.
interface ScanDecision {
  valid: boolean;
  code?: string;
  reason?: string;
}

interface BlockedScan {
  sku: string;
  reason: string;
  code?: string;
}

export function ScannerPage() {
  const [active, setActive] = useState(false);
  const [scans, setScans] = useState<string[]>([]);
  const [pending, setPending] = useState(0);
  const [blocked, setBlocked] = useState<BlockedScan | null>(null);
  const [manualSku, setManualSku] = useState("");
  const api = useApi();

  const accept = useCallback((sku: string) => {
    setBlocked(null);
    setScans((prev) => (prev[0] === sku ? prev : [sku, ...prev].slice(0, 20)));
  }, []);

  const handleScan = useCallback(
    async (sku: string) => {
      const payload: ScanPayload = { sku, zoneId: "", scannedAt: Date.now() };

      try {
        const decision = await api.post<ScanDecision>("/scan", payload);
        // A mis-scan is rejected by the server — block it and surface the reason.
        if (decision && decision.valid === false) {
          setBlocked({ sku, reason: decision.reason ?? "Scan rejected", code: decision.code });
          return;
        }
        accept(sku);
      } catch {
        // Offline: queue for later sync and optimistically accept.
        await enqueueScan(payload);
        setPending(await pendingCount());
        accept(sku);
      }
    },
    [api, accept],
  );

  const submitManual = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const sku = manualSku.trim();
      if (!sku) return;
      void handleScan(sku);
      setManualSku("");
    },
    [manualSku, handleScan],
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

  return (
    <div>
      <div className="page-head">
        <div>
          <div className="page-title">Scanner</div>
          <div className="page-sub">Camera + WebHID barcode scanning — P2</div>
        </div>
        <div className="flex items-center gap-2">
          {pending > 0 && (
            <span className="rounded bg-amber-100 px-2 py-1 text-xs text-amber-800">
              {pending} queued offline
            </span>
          )}
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

      {/* Manual SKU entry — fallback when no camera/scanner, and the e2e scan hook. */}
      <form onSubmit={submitManual} className="mt-4 flex max-w-md gap-2">
        <input
          data-testid="manual-sku-input"
          value={manualSku}
          onChange={(e) => setManualSku(e.target.value)}
          placeholder="Enter SKU"
          className="flex-1 rounded border border-slate-300 px-3 py-2 font-mono text-sm"
        />
        <button data-testid="manual-scan-btn" type="submit">
          Scan
        </button>
      </form>

      {blocked && (
        <div
          data-testid="scan-blocked"
          role="alert"
          className="mt-4 max-w-md rounded-lg border border-red-300 bg-red-50 px-4 py-3"
        >
          <div className="font-semibold text-red-800">Scan blocked — {blocked.sku}</div>
          <div data-testid="scan-blocked-reason" className="text-sm text-red-700">
            {blocked.reason}
          </div>
        </div>
      )}

      <ul className="mt-4 space-y-1" data-testid="accepted-scans">
        {scans.map((sku, i) => (
          <li key={`${sku}-${i}`} className="rounded bg-slate-100 px-3 py-2 font-mono text-sm">
            {sku}
          </li>
        ))}
      </ul>
    </div>
  );
}
