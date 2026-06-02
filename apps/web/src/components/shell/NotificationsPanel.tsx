import { useEffect, useRef } from "react";
import type { NotificationItem } from "../../lib/useNotificationCenter";

interface Props {
  items: NotificationItem[];
  onClose: () => void;
}

const KIND_LABEL: Record<NotificationItem["kind"], string> = {
  assigned: "Assigned",
  deadline: "Due soon",
};

export function NotificationsPanel({ items, onClose }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  // Close on outside click.
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [onClose]);

  return (
    <div
      ref={ref}
      role="dialog"
      aria-label="Notifications"
      style={{
        position: "absolute",
        top: "calc(100% + 8px)",
        right: 0,
        width: 320,
        maxHeight: 400,
        overflowY: "auto",
        zIndex: 50,
      }}
      className="overflow-hidden rounded-lg border border-border bg-surface shadow-lg"
    >
      <div className="border-b border-border px-3 py-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        Notifications
      </div>
      {items.length === 0 ? (
        <div className="px-3 py-6 text-center text-sm text-muted-foreground">No notifications</div>
      ) : (
        <ul>
          {items.map((n) => (
            <li
              key={n.id}
              className={`flex items-start gap-2 border-b border-border px-3 py-2 text-sm last:border-b-0 ${
                n.read ? "" : "bg-primary/5"
              }`}
            >
              <span
                className={`mt-1 h-2 w-2 shrink-0 rounded-full ${
                  n.kind === "deadline" ? "bg-warning" : "bg-primary"
                }`}
              />
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium text-foreground">{n.title}</p>
                <p className="text-xs text-muted-foreground">
                  {KIND_LABEL[n.kind]} · {new Date(n.ts).toLocaleString()}
                </p>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
