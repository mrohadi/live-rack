import { BrowserMultiFormatReader, type IScannerControls } from "@zxing/browser";
import { useEffect, useRef } from "react";

interface UseBarcodeScannerOptions {
  onScan: (sku: string) => void;
  active: boolean;
  deviceId?: string;
}

export function useBarcodeScanner({ onScan, active, deviceId }: UseBarcodeScannerOptions) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const onScanRef = useRef(onScan);
  onScanRef.current = onScan;

  useEffect(() => {
    if (!active || !videoRef.current) return;

    const reader = new BrowserMultiFormatReader();
    let controls: IScannerControls | undefined;

    reader
      .decodeFromVideoDevice(deviceId, videoRef.current, (result) => {
        if (result) onScanRef.current(result.getText());
      })
      .then((c) => {
        controls = c;
      })
      .catch((err) => console.error("camera start failed", err));

    return () => controls?.stop();
  }, [active, deviceId]);

  return videoRef;
}
