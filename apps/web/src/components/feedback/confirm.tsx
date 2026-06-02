import { useCallback, useRef, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { ConfirmContext, type ConfirmFn, type ConfirmOptions } from "./confirm-context";

/** Wrap the app to enable useConfirm(). Renders a portal-mounted modal. */
export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [opts, setOpts] = useState<ConfirmOptions | null>(null);
  const resolver = useRef<((ok: boolean) => void) | null>(null);

  const confirm = useCallback<ConfirmFn>((options) => {
    setOpts(options);
    return new Promise<boolean>((resolve) => {
      resolver.current = resolve;
    });
  }, []);

  const settle = useCallback((ok: boolean) => {
    resolver.current?.(ok);
    resolver.current = null;
    setOpts(null);
  }, []);

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      {opts &&
        createPortal(
          <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
            onMouseDown={() => settle(false)}
          >
            <div
              role="dialog"
              aria-modal="true"
              aria-label={opts.title}
              onMouseDown={(e) => e.stopPropagation()}
              className="w-full max-w-sm rounded-lg border border-border bg-surface p-5 shadow-lg"
            >
              <h2 className="text-base font-semibold text-foreground">{opts.title}</h2>
              {opts.message && <p className="mt-2 text-sm text-muted-foreground">{opts.message}</p>}
              <div className="mt-5 flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => settle(false)}
                  className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition hover:bg-muted"
                >
                  {opts.cancelLabel ?? "Cancel"}
                </button>
                <button
                  type="button"
                  autoFocus
                  onClick={() => settle(true)}
                  className={`rounded-md px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90 ${
                    opts.destructive ? "bg-red-600" : "bg-primary"
                  }`}
                >
                  {opts.confirmLabel ?? "Confirm"}
                </button>
              </div>
            </div>
          </div>,
          document.body,
        )}
    </ConfirmContext.Provider>
  );
}
