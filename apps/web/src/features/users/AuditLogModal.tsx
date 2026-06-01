import { auditLabel, relativeTime, useAudit } from "./useUsers";

interface AuditLogModalProps {
  onClose: () => void;
}

/** Modal listing the org's recent audit-trail entries. */
export function AuditLogModal({ onClose }: AuditLogModalProps) {
  const audit = useAudit(undefined, 50);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Audit log"
    >
      <div className="flex max-h-[80vh] w-full max-w-lg flex-col rounded-lg border border-border bg-surface shadow-lg">
        <div className="flex items-center justify-between border-b border-border px-5 py-3">
          <h2 className="text-base font-semibold text-foreground">Audit log</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded p-1 text-muted-foreground hover:text-foreground"
          >
            ✕
          </button>
        </div>
        <div className="flex-1 overflow-auto p-2">
          {audit.isLoading ? (
            <div className="p-4 text-sm text-muted-foreground">Loading…</div>
          ) : (audit.data ?? []).length === 0 ? (
            <div className="p-4 text-sm text-muted-foreground">No audit entries yet.</div>
          ) : (
            <ul className="divide-y divide-border">
              {audit.data!.map((e, i) => (
                <li key={i} className="flex items-center justify-between gap-3 px-3 py-2 text-sm">
                  <div className="min-w-0">
                    <div className="font-medium text-foreground">{auditLabel(e.action)}</div>
                    <div className="truncate text-xs text-muted-foreground">
                      {e.resource_type}
                      {e.resource_id ? ` · ${e.resource_id}` : ""}
                    </div>
                  </div>
                  <span className="shrink-0 font-mono text-xs text-muted-foreground">
                    {relativeTime(e.ts)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
