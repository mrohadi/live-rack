import type { PickLine } from "./types";

/** A compact map of the route: stops plotted by zone coords, joined in order. */
export function PickRoute({ lines }: { lines: PickLine[] }) {
  const stops = [...lines].sort((a, b) => a.seq - b.seq).filter((l) => l.zone_id);
  if (stops.length === 0) {
    return <div className="text-xs text-muted-foreground">No mapped locations for this route.</div>;
  }

  // Normalise coordinates into a 0–100 viewBox with a small margin.
  const xs = stops.map((s) => s.zone_x);
  const ys = stops.map((s) => s.zone_y);
  const minX = Math.min(0, ...xs);
  const maxX = Math.max(100, ...xs);
  const minY = Math.min(0, ...ys);
  const maxY = Math.max(100, ...ys);
  const nx = (v: number) => 6 + ((v - minX) / (maxX - minX || 1)) * 88;
  const ny = (v: number) => 6 + ((v - minY) / (maxY - minY || 1)) * 88;

  const path = stops
    .map((s, i) => `${i === 0 ? "M" : "L"} ${nx(s.zone_x)} ${ny(s.zone_y)}`)
    .join(" ");

  return (
    <svg
      viewBox="0 0 100 100"
      className="h-44 w-full rounded-lg border border-border bg-surface"
      role="img"
      aria-label="Pick route map"
    >
      <path
        d={path}
        fill="none"
        stroke="currentColor"
        strokeWidth="0.8"
        className="text-primary/50"
      />
      {stops.map((s, i) => (
        <g key={s.id}>
          <circle
            cx={nx(s.zone_x)}
            cy={ny(s.zone_y)}
            r="3"
            className={s.status === "pending" ? "fill-primary" : "fill-emerald-500"}
          />
          <text
            x={nx(s.zone_x)}
            y={ny(s.zone_y) + 1}
            textAnchor="middle"
            className="fill-white text-[3px]"
          >
            {i + 1}
          </text>
        </g>
      ))}
    </svg>
  );
}
