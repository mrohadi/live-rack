import { useCallback, useMemo, useRef, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { ToastContext, type ToastApi, type ToastVariant } from "./toast-context";

interface Toast {
  id: number;
  message: string;
  variant: ToastVariant;
}

const DURATION_MS = 4000;

const VARIANT_STYLES: Record<ToastVariant, string> = {
  success: "border-green-500/40 bg-green-50 text-green-800",
  error: "border-red-500/40 bg-red-50 text-red-800",
  info: "border-border bg-surface text-foreground",
};

const VARIANT_ICON: Record<ToastVariant, string> = {
  success: "✓",
  error: "✕",
  info: "ℹ",
};

/** Wrap the app to enable useToast(). Renders a portal-mounted stack. */
export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextId = useRef(1);

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const push = useCallback(
    (message: string, variant: ToastVariant = "info") => {
      const id = nextId.current++;
      setToasts((prev) => [...prev, { id, message, variant }]);
      window.setTimeout(() => remove(id), DURATION_MS);
    },
    [remove],
  );

  const api = useMemo<ToastApi>(
    () => ({
      push,
      success: (m) => push(m, "success"),
      error: (m) => push(m, "error"),
      info: (m) => push(m, "info"),
    }),
    [push],
  );

  return (
    <ToastContext.Provider value={api}>
      {children}
      {createPortal(
        <div className="pointer-events-none fixed left-1/2 top-5 z-50 flex -translate-x-1/2 flex-col items-center gap-2">
          {toasts.map((t) => (
            <div
              key={t.id}
              role="status"
              className={`pointer-events-auto flex min-w-[320px] items-center gap-3 rounded-lg border px-4 py-3 text-sm font-medium shadow-lg ${VARIANT_STYLES[t.variant]}`}
            >
              <span aria-hidden>{VARIANT_ICON[t.variant]}</span>
              <span>{t.message}</span>
              <button
                type="button"
                aria-label="dismiss"
                onClick={() => remove(t.id)}
                className="ml-1 opacity-60 hover:opacity-100"
              >
                ×
              </button>
            </div>
          ))}
        </div>,
        document.body,
      )}
    </ToastContext.Provider>
  );
}
