import type { InventoryRow } from "../inventory/types";
import type { ViewMode, Zone } from "./types";
import { fillRatio, rgba, slotStatus } from "./zoneMath";

interface Props {
  zones: Zone[];
  items: InventoryRow[];
  view: ViewMode;
  selectedId: string | null;
  onSelect: (id: string) => void;
}

const HEAT_ACCENT = "#2563eb";

const LEGEND = [
  { label: "Apparel", color: "#2563eb" },
  { label: "Electronics / Cold", color: "#0891b2" },
  { label: "Home / Showroom", color: "#16a34a" },
  { label: "Receiving", color: "#7c3aed" },
  { label: "Returns", color: "#d97706" },
  { label: "Outbound", color: "#dc2626" },
  { label: "Bulk / Staging", color: "#5b6577" },
];

export function ZoneMapView({ zones, items, view, selectedId, onSelect }: Props) {
  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div
        className="relative flex-1 overflow-hidden rounded-lg border border-border bg-background"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(0,0,0,0.04) 1px, transparent 1px), linear-gradient(to bottom, rgba(0,0,0,0.04) 1px, transparent 1px)",
          backgroundSize: "24px 24px",
        }}
      >
        {zones.map((z) => (
          <ZoneBox
            key={z.id}
            zone={z}
            items={items.filter((it) => it.zone_id === z.id)}
            view={view}
            selected={selectedId === z.id}
            onSelect={() => onSelect(z.id)}
          />
        ))}

        {/* SKU location pins for the selected zone (Items view) */}
        {view === "items" &&
          (() => {
            const zn = zones.find((z) => z.id === selectedId);
            if (!zn) return null;
            const cols = Math.max(4, Math.round(zn.width / 4));
            const here = items.filter((it) => it.zone_id === zn.id).slice(0, 3);
            return here.map((it, i) => {
              const idx = i * 3 + 1;
              const col = String.fromCharCode(65 + (idx % cols));
              const row = String(Math.floor(idx / cols) + 1).padStart(2, "0");
              return (
                <div
                  key={it.sku}
                  className="pointer-events-none absolute z-10 -translate-x-1/2 -translate-y-full whitespace-nowrap rounded-md border border-border bg-surface px-2 py-1 font-mono text-[10px] text-foreground shadow-sm"
                  style={{ left: `${zn.x + (zn.width * (i + 1)) / 4}%`, top: `${zn.y}%` }}
                >
                  <span
                    className="mr-1 inline-block h-1.5 w-1.5 rounded-full align-middle"
                    style={{ background: zn.color }}
                  />
                  {it.sku} · {zn.id}-{row}
                  {col}
                </div>
              );
            });
          })()}

        {/* Aisle marker */}
        <div className="pointer-events-none absolute left-1/2 top-[74%] -translate-x-1/2 font-mono text-[10px] tracking-[0.2em] text-muted-foreground">
          —— MAIN AISLE ——
        </div>
        {/* Compass / scale */}
        <div className="absolute right-2.5 top-2.5 rounded-md border border-border bg-surface px-2 py-1 font-mono text-[11px] text-muted-foreground">
          N ↑ · 1:120
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-x-4 gap-y-1.5 text-xs text-muted-foreground">
        {LEGEND.map((l) => (
          <span key={l.label} className="inline-flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rounded-sm" style={{ background: l.color }} />
            {l.label}
          </span>
        ))}
      </div>
    </div>
  );
}

function ZoneBox({
  zone,
  items,
  view,
  selected,
  onSelect,
}: {
  zone: Zone;
  items: InventoryRow[];
  view: ViewMode;
  selected: boolean;
  onSelect: () => void;
}) {
  const fill = fillRatio(zone);
  const pct = Math.round(fill * 100);
  const heatBg = view === "heat" ? rgba(HEAT_ACCENT, 0.12 + fill * 0.78) : rgba(zone.color, 0.06);

  return (
    <button
      type="button"
      data-testid="zone-box"
      onClick={onSelect}
      className={`absolute flex flex-col overflow-hidden rounded-lg border-2 p-2 text-left transition ${
        selected ? "ring-2 ring-offset-1" : ""
      }`}
      style={{
        left: `${zone.x}%`,
        top: `${zone.y}%`,
        width: `${zone.width}%`,
        height: `${zone.height}%`,
        borderColor: zone.color,
        background: heatBg,
        color: zone.color,
      }}
    >
      <div className="text-xs font-semibold">{zone.name}</div>
      <div className="font-mono text-[10px] opacity-80">
        {zone.items ?? 0}/{zone.capacity ?? 0} · {pct}%
      </div>

      {view === "items" && <RackGrid zone={zone} items={items} fill={fill} />}
    </button>
  );
}

function RackGrid({ zone, items, fill }: { zone: Zone; items: InventoryRow[]; fill: number }) {
  const cols = Math.max(4, Math.round(zone.width / 4));
  const rows = Math.max(3, Math.round(zone.height / 6));
  const total = cols * rows;
  const filled = Math.min(total, Math.round(total * fill));

  return (
    <div
      className="mt-1.5 grid flex-1 gap-0.5"
      style={{
        gridTemplateColumns: `repeat(${cols}, 1fr)`,
        gridTemplateRows: `repeat(${rows}, 1fr)`,
      }}
    >
      {Array.from({ length: total }).map((_, i) => {
        const isFilled = i < filled;
        const item = isFilled && items.length > 0 ? items[i % items.length] : null;
        const status = item ? slotStatus(item.qty) : "ok";
        const bg = !isFilled
          ? "transparent"
          : status === "out"
            ? "#dc2626"
            : status === "low"
              ? "#d97706"
              : zone.color;
        return (
          <span
            key={i}
            className="rounded-[2px]"
            style={{
              background: bg,
              border: isFilled ? "none" : `1px solid ${rgba(zone.color, 0.3)}`,
            }}
          />
        );
      })}
    </div>
  );
}
