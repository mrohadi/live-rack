import type { ZoneCanvasProps } from "../types";

export function SvgZoneCanvas({ zones, selectedIds, onSelect }: ZoneCanvasProps) {
  return (
    <svg width="100%" height="100%" style={{ display: "block" }} aria-label="Zone map">
      {zones.map((zone) => (
        <rect
          key={zone.id}
          data-zone-id={zone.id}
          data-selected={selectedIds.includes(zone.id) ? "true" : "false"}
          x={zone.x}
          y={zone.y}
          width={zone.width}
          height={zone.height}
          fill={zone.color}
          stroke={selectedIds.includes(zone.id) ? "#fff" : "transparent"}
          strokeWidth={selectedIds.includes(zone.id) ? 2 : 0}
          rx={4}
          style={{ cursor: "pointer" }}
          onClick={() => onSelect([zone.id])}
        />
      ))}
    </svg>
  );
}
