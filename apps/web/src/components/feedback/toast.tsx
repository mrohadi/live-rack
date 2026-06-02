import { createPortal } from "react-dom";
import { useCallback, useRef, useState } from "react";
import { ToastContext, type ToastVariant } from "./toast-context";

interface Toast {
  id: number;
  msg: string;
  variant: ToastVariant;
}

const VARIANT_STYLES: Record<ToastVariant, string> = {
  success: "bg-green-700 text-white",
  error: "bg-destructive text-white",
  info: "bg-primary text-white",
};

const VARIANT_ICON: Record<ToastVariant, string> = {
  success: "✓",
  error: "✕",
  info: "ℹ",
};

let _id = 0;

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const timers = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map());

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
    timers.current.delete(id);
  }, []);

  const push = useCallback(
    (msg: string, variant: ToastVariant) => {
      const id = ++_id;
      setToasts((prev) => [...prev, { id, msg, variant }]);
      timers.current.set(
        id,
        setTimeout(() => remove(id), 4000),
      );
    },
    [remove],
  );

  const api = {
    success: (msg: string) => push(msg, "success"),
    error: (msg: string) => push(msg, "error"),
    info: (msg: string) => push(msg, "info"),
  };

  return (
    <ToastContext.Provider value={api}>
      {children}
      {createPortal(
        <div className="fixed bottom-4 right-4 z-[9999] flex flex-col gap-2">
          {toasts.map((t) => (
            <div
              key={t.id}
              role="status"
              className={`flex items-center gap-2 rounded-lg px-4 py-2.5 text-sm shadow-lg ${VARIANT_STYLES[t.variant]}`}
            >
              <span className="font-bold">{VARIANT_ICON[t.variant]}</span>
              <span>{t.msg}</span>
              <button
                type="button"
                aria-label="dismiss"
                onClick={() => remove(t.id)}
                className="ml-2 opacity-70 hover:opacity-100"
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
