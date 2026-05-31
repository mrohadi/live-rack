import { sparkPoints } from "../dashboard/useSales";
import { barWidthPct, type ZonePerf } from "./useAnalytics";

interface ZonePerfBarsProps {
  zones: ZonePerf[];
}

/** Horizontal scan-volume bars per zone with a recent activity sparkline. */
export function ZonePerfBars({ zones }: ZonePerfBarsProps) {
  if (zones.length === 0) {
    return <div className="text-sm text-muted-foreground">No zone activity yet.</div>;
  }
  return (
    <div className="flex flex-col gap-2" data-testid="zone-perf">
      {zones.map((z) => (
        <div key={z.zone_id} className="flex items-center gap-3">
          <span className="w-28 truncate font-mono text-xs text-foreground">{z.zone_id.slice(0, 8)}</span>
          <div className="h-2 flex-1 rounded bg-border">
            <div
              className="h-2 rounded bg-primary"
              style={{ width: `${barWidthPct(z.scans, zones)}%` }}
            />
          </div>
          <span className="w-14 text-right font-mono text-xs text-foreground">{z.scans}</span>
          <span
            className="w-12 text-right font-mono text-[11px] text-danger"
            title="invalid scans"
          >
            {z.invalid > 0 ? `⚠ ${z.invalid}` : ""}
          </span>
          <svg
            role="img"
            aria-label={`${z.zone_id} activity`}
            viewBox="0 0 100 28"
            preserveAspectRatio="none"
            className="h-7 w-24 text-primary"
          >
            <polyline points={sparkPoints(z.spark, 100, 28)} fill="none" stroke="currentColor" strokeWidth="1.4" />
          </svg>
        </div>
      ))}
    </div>
  );
}
