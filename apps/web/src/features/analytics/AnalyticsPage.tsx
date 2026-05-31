import { Heatmap } from "./Heatmap";
import { useHeatmap } from "./useAnalytics";

export function AnalyticsPage() {
  const { data, isLoading } = useHeatmap();

  return (
    <div className="flex h-full flex-col">
      <header className="border-b border-border px-4 py-3">
        <h1 className="text-lg font-semibold text-foreground">Analytics</h1>
        <p className="text-xs text-muted-foreground">Scan activity · last 7 days</p>
      </header>

      <div className="flex-1 overflow-auto p-4">
        <div className="rounded-lg border border-border bg-surface p-4">
          <div className="mb-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Activity heatmap · 7×24
          </div>
          {isLoading || !data ? (
            <div className="text-sm text-muted-foreground">Loading heatmap…</div>
          ) : (
            <Heatmap data={data} />
          )}
        </div>
      </div>
    </div>
  );
}
