import { useEffect } from "react";
import { useAuth } from "react-oidc-context";
import { openTaskNotificationSocket, type TaskNotification } from "./ws";

// useTaskNotifications subscribes to live task.notified events. Pass a stable
// handler (useCallback) to avoid reconnecting on every render.
export function useTaskNotifications(onNotify: (n: TaskNotification) => void): void {
  const auth = useAuth();
  const token = auth.user?.access_token ?? null;
  useEffect(() => {
    return openTaskNotificationSocket(async () => token, onNotify);
  }, [token, onNotify]);
}
