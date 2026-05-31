import { HOURS_IN_DAY, WEEKDAYS, heatColor, type HeatmapResponse } from "./useAnalytics";

interface HeatmapProps {
  data: HeatmapResponse;
  accent?: string;
}

/** 7x24 activity heatmap. Row 0 = Monday; columns are hours 0..23. */
export function Heatmap({ data, accent = "var(--accent)" }: HeatmapProps) {
  return (
    <div className="flex flex-col gap-1" data-testid="heatmap">
      <div className="grid" style={{ gridTemplateColumns: `36px repeat(${HOURS_IN_DAY},1fr)` }}>
        <div />
        {Array.from({ length: HOURS_IN_DAY }).map((_, h) => (
          <div key={h} className="text-center font-mono text-[9.5px] text-[var(--text-3)]">
            {h % 6 === 0 ? h : ""}
          </div>
        ))}
      </div>
      {data.grid.map((row, di) => (
        <div
          key={WEEKDAYS[di]}
          className="grid items-center"
          style={{ gridTemplateColumns: `36px repeat(${HOURS_IN_DAY},1fr)` }}
        >
          <div className="font-mono text-[10px] text-[var(--text-3)]">{WEEKDAYS[di]}</div>
          {row.map((v, h) => (
            <div
              key={h}
              className="aspect-square rounded-[2px]"
              title={`${WEEKDAYS[di]} ${h}:00 · ${v}`}
              style={{ background: heatColor(v, data.max, accent) }}
            />
          ))}
        </div>
      ))}
    </div>
  );
}
