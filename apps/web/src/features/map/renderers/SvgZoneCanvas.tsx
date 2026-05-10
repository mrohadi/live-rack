import type { ZoneCanvasProps } from "../types";

export function SvgZoneCanvas({ zones, selectedId, onSelect }: ZoneCanvasProps) {
  return (
    <svg width="100%" height="100%" style={{ display: "block" }} aria-label="Zone map">
      {zones.map((zone) => (
        <rect
          key={zone.id}
          data-zone-id={zone.id}
          data-selected={zone.id === selectedId ? "true" : "false"}
          x={zone.x}
          y={zone.y}
          width={zone.width}
          height={zone.height}
          fill={zone.color}
          stroke={zone.id === selectedId ? "#fff" : "transparent"}
          strokeWidth={zone.id === selectedId ? 2 : 0}
          rx={4}
          style={{ cursor: "pointer" }}
          onClick={() => onSelect(zone.id)}
        />
      ))}
    </svg>
  );
}
