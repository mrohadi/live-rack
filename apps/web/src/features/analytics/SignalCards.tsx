import type { Recommendation } from "../../lib/ws";

interface SignalCardsProps {
  recommendations: Recommendation[];
  onApply: (rec: Recommendation) => void;
  applyingId?: string;
}

/** Live recommendation cards from external signals, each with an Apply button
 *  that turns the suggested action into a task. */
export function SignalCards({ recommendations, onApply, applyingId }: SignalCardsProps) {
  if (recommendations.length === 0) {
    return <div className="text-sm text-muted-foreground">No active signals.</div>;
  }
  return (
    <div className="flex flex-col gap-2" data-testid="signal-cards">
      {recommendations.map((rec) => (
        <div
          key={rec.id}
          className="flex items-start justify-between gap-3 rounded-md border border-border bg-surface p-3"
        >
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className="rounded bg-border px-1.5 py-0.5 font-mono text-[10px] uppercase text-muted-foreground">
                {rec.kind}
              </span>
              <span className="truncate text-sm font-medium text-foreground">{rec.title}</span>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">{rec.rationale}</p>
            <p className="mt-1 text-xs text-foreground">→ {rec.suggested_task}</p>
          </div>
          <button
            type="button"
            onClick={() => onApply(rec)}
            disabled={applyingId === rec.id}
            className="shrink-0 rounded bg-primary px-3 py-1 text-xs font-medium text-white disabled:opacity-50"
          >
            {applyingId === rec.id ? "Applying…" : "Apply"}
          </button>
        </div>
      ))}
    </div>
  );
}
