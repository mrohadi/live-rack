import { Heatmap } from "./Heatmap";
import { SignalCards } from "./SignalCards";
import { ZonePerfBars } from "./ZonePerfBars";
import { useHeatmap, useZonePerf } from "./useAnalytics";
import { useApplyRecommendation, useRecommendations } from "./useRecommendations";

export function AnalyticsPage() {
  const heatmap = useHeatmap();
  const zones = useZonePerf();
  const recommendations = useRecommendations();
  const apply = useApplyRecommendation();

  return (
    <div className="flex h-full flex-col">
      <header className="border-b border-border px-4 py-3">
        <h1 className="text-lg font-semibold text-foreground">Analytics</h1>
        <p className="text-xs text-muted-foreground">Scan activity · last 7 days</p>
      </header>

      <div className="flex-1 space-y-4 overflow-auto p-4">
        <div className="rounded-lg border border-border bg-surface p-4">
          <div className="mb-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Signals · recommendations
          </div>
          <SignalCards
            recommendations={recommendations}
            onApply={(rec) => apply.mutate(rec)}
            applyingId={apply.isPending ? apply.variables?.id : undefined}
          />
        </div>

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
