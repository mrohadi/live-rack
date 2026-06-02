import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { Select } from "../../components/ui/Select";
import { useQueryClient } from "@tanstack/react-query";
import { useZones } from "../map/useZones";
import { useCurrentStore } from "../map/useCurrentStore";
import { inventoryKeys } from "./useInventory";
import {
  useCompleteCount,
  useSetCountLine,
  useStartCount,
  type CompleteCountResult,
  type CountSession,
} from "./useCounts";

interface Props {
  onClose: () => void;
}

export function CycleCountModal({ onClose }: Props) {
  const storeId = useCurrentStore();
  const { data: zones = [] } = useZones(storeId);
  const start = useStartCount(storeId);
  const toast = useToast();

  const [zoneId, setZoneId] = useState("");
  const [session, setSession] = useState<CountSession | null>(null);
  const [result, setResult] = useState<CompleteCountResult | null>(null);

  const zoneOptions = zones.map((z) => ({ value: z.id, label: z.name }));

  const begin = () => {
    if (!zoneId) return;
    start.mutate(zoneId, {
      onSuccess: (s) => setSession(s),
      onError: () => toast.error("Failed to start count"),
    });
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Cycle count"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="flex max-h-[85vh] w-full max-w-lg flex-col rounded-lg border border-border bg-surface p-5 shadow-lg">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-base font-semibold text-foreground">Cycle count</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            ✕
          </button>
        </div>

        {result ? (
          <ResultView result={result} onClose={onClose} />
        ) : session ? (
          <CountEntry storeId={storeId} session={session} onComplete={setResult} />
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Pick a zone to count. Quantities are entered blind; the system compares them on
              completion.
            </p>
            <label className="block text-sm">
              <span className="mb-1.5 block text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Zone *
              </span>
              <Select
                value={zoneId}
                onChange={setZoneId}
                options={zoneOptions}
                placeholder="Select zone…"
              />
            </label>
            <div className="flex justify-end gap-2 pt-1">
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg border border-border px-3 py-1.5 text-sm text-foreground transition hover:bg-muted"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={begin}
                disabled={!zoneId || start.isPending}
                className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
              >
                {start.isPending ? "Starting…" : "Start count"}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function CountEntry({
  storeId,
  session,
  onComplete,
}: {
  storeId: string;
  session: CountSession;
  onComplete: (r: CompleteCountResult) => void;
}) {
  const setLine = useSetCountLine(storeId, session.id);
  const complete = useCompleteCount(storeId, session.id);
  const qc = useQueryClient();
  const toast = useToast();
  const [counts, setCounts] = useState<Record<string, string>>({});

  const save = (sku: string, raw: string) => {
    setCounts((c) => ({ ...c, [sku]: raw }));
    const n = Math.max(0, Math.floor(Number(raw)));
    if (raw === "" || Number.isNaN(n)) return;
    setLine.mutate({ sku, counted_qty: n });
  };

  const finish = () => {
    complete.mutate(undefined, {
      onSuccess: (r) => {
        toast.success(`Reconciled ${r.reconciled} item(s)`);
        void qc.invalidateQueries({ queryKey: inventoryKeys.list(storeId) });
        onComplete(r);
      },
      onError: () => toast.error("Failed to complete count"),
    });
  };

  if (session.lines.length === 0) {
    return <p className="text-sm text-muted-foreground">No stock in this zone to count.</p>;
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <p className="mb-2 text-xs text-muted-foreground">
        Enter the physical count for each SKU. Blank lines are left untouched.
      </p>
      <div className="min-h-0 flex-1 overflow-auto rounded-lg border border-border">
        <table className="w-full text-sm">
          <thead className="sticky top-0 bg-muted/40 text-left text-muted-foreground">
            <tr>
              <th className="px-3 py-1.5 font-medium">SKU</th>
              <th className="px-3 py-1.5 text-right font-medium">Counted</th>
            </tr>
          </thead>
          <tbody>
            {session.lines.map((l) => (
              <tr key={l.sku} className="border-t border-border">
                <td className="px-3 py-1.5 font-mono text-xs">{l.sku}</td>
                <td className="px-3 py-1.5 text-right">
                  <input
                    type="number"
                    min={0}
                    value={counts[l.sku] ?? ""}
                    onChange={(e) => save(l.sku, e.target.value)}
                    placeholder="—"
                    className="w-20 rounded border border-border bg-background px-2 py-0.5 text-right text-sm outline-none focus:border-primary"
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="flex justify-end pt-3">
        <button
          type="button"
          onClick={finish}
          disabled={complete.isPending}
          className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
        >
          {complete.isPending ? "Reconciling…" : "Complete & reconcile"}
        </button>
      </div>
    </div>
  );
}

function ResultView({ result, onClose }: { result: CompleteCountResult; onClose: () => void }) {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <p className="mb-3 text-sm text-foreground">
        Count complete — <span className="font-semibold">{result.reconciled}</span> SKU(s) adjusted.
      </p>
      {result.variances.length === 0 ? (
        <p className="text-sm text-muted-foreground">No variances. Stock matched the system.</p>
      ) : (
        <div className="min-h-0 flex-1 overflow-auto rounded-lg border border-border">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-muted/40 text-left text-muted-foreground">
              <tr>
                <th className="px-3 py-1.5 font-medium">SKU</th>
                <th className="px-3 py-1.5 text-right font-medium">System</th>
                <th className="px-3 py-1.5 text-right font-medium">Counted</th>
                <th className="px-3 py-1.5 text-right font-medium">Variance</th>
              </tr>
            </thead>
            <tbody>
              {result.variances.map((v) => (
                <tr key={v.sku} className="border-t border-border">
                  <td className="px-3 py-1.5 font-mono text-xs">{v.sku}</td>
                  <td className="px-3 py-1.5 text-right text-muted-foreground">{v.system_qty}</td>
                  <td className="px-3 py-1.5 text-right text-foreground">{v.counted_qty}</td>
                  <td
                    className={`px-3 py-1.5 text-right font-semibold ${
                      v.variance < 0 ? "text-destructive" : "text-primary"
                    }`}
                  >
                    {v.variance > 0 ? `+${v.variance}` : v.variance}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <div className="flex justify-end pt-3">
        <button
          type="button"
          onClick={onClose}
          className="rounded-lg bg-primary px-4 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
        >
          Done
        </button>
      </div>
    </div>
  );
}
