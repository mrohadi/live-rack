import { Heatmap } from "./Heatmap";
import { ZonePerfBars } from "./ZonePerfBars";
import { useHeatmap, useZonePerf } from "./useAnalytics";

export function AnalyticsPage() {
  const heatmap = useHeatmap();
  const zones = useZonePerf();

  return (
    <div className="flex h-full flex-col">
      <header className="border-b border-border px-4 py-3">
        <h1 className="text-lg font-semibold text-foreground">Analytics</h1>
        <p className="text-xs text-muted-foreground">Scan activity · last 7 days</p>
      </header>

      <div className="flex-1 space-y-4 overflow-auto p-4">
        <div className="rounded-lg border border-border bg-surface p-4">
          <div className="mb-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Zone performance · scans
          </div>
          {zones.isLoading || !zones.data ? (
            <div className="text-sm text-muted-foreground">Loading zones…</div>
          ) : (
            <ZonePerfBars zones={zones.data.zones} />
          )}
        </div>

        <div className="rounded-lg border border-border bg-surface p-4">
          <div className="mb-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Activity heatmap · 7×24
          </div>
          {heatmap.isLoading || !heatmap.data ? (
            <div className="text-sm text-muted-foreground">Loading heatmap…</div>
          ) : (
            <Heatmap data={heatmap.data} />
          )}
        </div>
      </div>
    </div>
  );
}
