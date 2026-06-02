import { useCallback, useMemo, useRef, useState } from "react";
import { useCapabilities } from "../features/users/useUsers";
import type { TaskNotification } from "./ws";

/** One stored notification with a stable id + read flag. */
export interface NotificationItem extends TaskNotification {
  id: string;
  read: boolean;
}

const MAX_ITEMS = 30;

/**
 * Keeps a capped, newest-first list of task notifications addressed to the
 * current user, plus an unread count. Notifications for other assignees are
 * dropped so the bell only reflects the caller's work.
 */
export function useNotificationCenter() {
  const { data: me } = useCapabilities();
  const myId = me?.user_id ?? null;
  const [items, setItems] = useState<NotificationItem[]>([]);
  const seq = useRef(0);

  const push = useCallback(
    (n: TaskNotification): boolean => {
      // Only surface notifications targeted at the current user.
      if (myId && n.assignee_id && n.assignee_id !== myId) return false;
      seq.current += 1;
      const item: NotificationItem = { ...n, id: `${n.task_id}-${seq.current}`, read: false };
      setItems((cur) => [item, ...cur].slice(0, MAX_ITEMS));
      return true;
    },
    [myId],
  );

  const markAllRead = useCallback(() => {
    setItems((cur) => cur.map((i) => (i.read ? i : { ...i, read: true })));
  }, []);

  const clear = useCallback(() => setItems([]), []);

  const unread = useMemo(() => items.reduce((n, i) => n + (i.read ? 0 : 1), 0), [items]);

  return { items, unread, push, markAllRead, clear };
}
