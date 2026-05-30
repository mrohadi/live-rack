import { useEffect } from "react";
import { useAuth } from "@clerk/clerk-react";
import { openScanSocket, type ScanRecorded } from "./ws";

// useScanStream subscribes to live scan.recorded events. Pass a stable handler (useCallback).
export function useScanStream(onEvent: (ev: ScanRecorded) => void): void {
  const { getToken } = useAuth();
  useEffect(() => {
    return openScanSocket(() => getToken(), onEvent);
  }, [getToken, onEvent]);
}
