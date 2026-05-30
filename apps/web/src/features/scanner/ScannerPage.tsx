import { useCallback, useState } from "react";
import { useBarcodeScanner } from "./useBarcodeScanner";
import { useZebraHID } from "./useZebraHID";

export function ScannerPage() {
  const [active, setActive] = useState(false);
  const [scans, setScans] = useState<string[]>([]);

  const handleScan = useCallback((sku: string) => {
    setScans((prev) => (prev[0] === sku ? prev : [sku, ...prev].slice(0, 20)));
  }, []);

  const videoRef = useBarcodeScanner({ onScan: handleScan, active });
  const { connect } = useZebraHID({ onScan: handleScan });

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
