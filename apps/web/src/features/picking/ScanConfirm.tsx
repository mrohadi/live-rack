import { useState } from "react";
import { useBarcodeScanner } from "../scanner/useBarcodeScanner";
import { useZebraHID } from "../scanner/useZebraHID";

interface Props {
  active: boolean;
  onScan: (code: string) => void;
}

/** Camera + Zebra HID capture panel for hands-free pick confirmation. */
export function ScanConfirm({ active, onScan }: Props) {
  const [hidConnected, setHidConnected] = useState(false);
  const videoRef = useBarcodeScanner({ onScan, active });
  const { connect } = useZebraHID({
    onScan,
    onConnect: () => setHidConnected(true),
    onDisconnect: () => setHidConnected(false),
  });

  return (
    <div className="space-y-2 rounded-lg border border-border bg-card p-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">
          Scan a SKU to confirm the matching stop
        </span>
        <button
          type="button"
          onClick={() => void connect()}
          className="rounded-md border border-border px-2 py-1 text-xs font-medium text-foreground"
        >
          {hidConnected ? "Scanner connected" : "Connect USB scanner"}
        </button>
      </div>
      {active && (
        <video
          ref={videoRef}
          className="h-40 w-full rounded-md bg-black object-cover"
          muted
          playsInline
        >
          <track kind="captions" />
        </video>
      )}
    </div>
  );
}
