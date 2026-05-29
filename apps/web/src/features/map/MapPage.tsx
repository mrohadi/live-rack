import { useState } from "react";
import { KonvaZoneCanvas } from "./renderers/KonvaZoneCanvas";
import type { Zone, ZoneUpdate, ViewMode } from "./types";

const seedZones: Zone[] = [
  {
    id: "z1",
    name: "Zone A",
    x: 40,
    y: 40,
    width: 200,
    height: 140,
    color: "#6366f1",
    type: "general",
    items: 140,
    capacity: 200,
  },
  {
    id: "z2",
    name: "Zone B",
    x: 280,
    y: 40,
    width: 200,
    height: 140,
    color: "#10b981",
    type: "frozen",
    items: 60,
    capacity: 200,
  },
  {
    id: "z3",
    name: "Zone C",
    x: 40,
    y: 220,
    width: 320,
    height: 160,
    color: "#f59e0b",
    type: "staging",
    items: 280,
    capacity: 320,
  },
];

export function MapPage() {
  const [zones, setZones] = useState<Zone[]>(seedZones);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [view, setView] = useState<ViewMode>("zones");

  const handleChange = (updates: ZoneUpdate[]) => {
    // TODO: persist via zones API (LR-103 endpoints). For now, update local state.
    setZones((prev) =>
      prev.map((z) => {
        const u = updates.find((upd) => upd.id === z.id);
        return u ? { ...z, ...u } : z;
      }),
    );
    // eslint-disable-next-line no-console
    console.log("zone updates:", updates);
  };

  const TABS: { label: string; value: ViewMode }[] = [
    { label: "Zones", value: "zones" },
    { label: "Heat", value: "heat" },
    { label: "Items", value: "items" },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <header
        style={{
          padding: "12px 16px",
          borderBottom: "1px solid #1f2937",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <div>
          <h1 style={{ margin: 0, fontSize: 18 }}>Map &amp; Zones</h1>
          <p style={{ margin: "4px 0 0", color: "#94a3b8", fontSize: 13 }}>
            Drag to move • drag corners to resize • shift-click for multi-select
          </p>
        </div>
        <div
          style={{ display: "flex", gap: 4, background: "#1e293b", borderRadius: 6, padding: 4 }}
        >
          {TABS.map((t) => (
            <button
              key={t.value}
              onClick={() => setView(t.value)}
              style={{
                padding: "4px 14px",
                borderRadius: 4,
                border: "none",
                cursor: "pointer",
                fontSize: 13,
                background: view === t.value ? "#334155" : "transparent",
                color: view === t.value ? "#f1f5f9" : "#94a3b8",
                fontWeight: view === t.value ? 600 : 400,
              }}
            >
              {t.label}
            </button>
          ))}
        </div>
      </header>
      <div style={{ flex: 1, position: "relative", background: "#0f172a" }}>
        <KonvaZoneCanvas
          zones={zones}
          selectedIds={selectedIds}
          onSelect={setSelectedIds}
          onChange={handleChange}
          viewMode={view}
        />
      </div>
    </div>
  );
}
