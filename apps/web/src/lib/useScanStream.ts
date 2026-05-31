import { useEffect } from "react";
import { useAuth } from "react-oidc-context";
import { openScanSocket, type ScanRecorded } from "./ws";

// useScanStream subscribes to live scan.recorded events. Pass a stable handler (useCallback).
export function useScanStream(onEvent: (ev: ScanRecorded) => void): void {
  const auth = useAuth();
  const token = auth.user?.access_token ?? null;
  useEffect(() => {
    return openScanSocket(async () => token, onEvent);
  }, [token, onEvent]);
}
