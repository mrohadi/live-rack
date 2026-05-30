import { useState } from "react";
import { KonvaZoneCanvas } from "./renderers/KonvaZoneCanvas";
import type { ViewMode, ZoneUpdate } from "./types";
import { useCurrentStore } from "./useCurrentStore";
import { useCreateZone, useUpdateZone, useZones } from "./useZones";
import { ZoneDetailSidebar } from "./ZoneDetailSidebar";

export function MapPage() {
  const storeId = useCurrentStore();
  const { data: zones = [], isLoading } = useZones(storeId);
  const createZone = useCreateZone(storeId);
  const updateZone = useUpdateZone(storeId);

  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [view, setView] = useState<ViewMode>("zones");
  const [showAddForm, setShowAddForm] = useState(false);
  const [newZoneName, setNewZoneName] = useState("");

  /** Merge drag/resize delta into the full zone, then PUT. */
  const handleChange = (updates: ZoneUpdate[]) => {
    updates.forEach((delta) => {
      const existing = zones.find((z) => z.id === delta.id);
      if (!existing) return;
      updateZone.mutate({ ...existing, ...delta });
    });
  };

  const handleAddZone = () => {
    if (!newZoneName.trim()) return;
    createZone.mutate(
      {
        name: newZoneName.trim(),
        type: "general",
        x: 40,
        y: 40,
        width: 200,
        height: 120,
        color: "#6366f1",
        capacity: 100,
      },
      {
        onSuccess: () => {
          setNewZoneName("");
          setShowAddForm(false);
        },
      },
    );
  };

  const TABS: { label: string; value: ViewMode }[] = [
    { label: "Zones", value: "zones" },
    { label: "Heat", value: "heat" },
    { label: "Items", value: "items" },
  ];

  const selectedZone =
    selectedIds.length === 1 ? (zones.find((z) => z.id === selectedIds[0]) ?? null) : null;

  if (isLoading) {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          height: "100%",
          color: "#94a3b8",
        }}
      >
        Loading zones…
      </div>
    );
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <header
        style={{
          padding: "12px 16px",
          borderBottom: "1px solid #1f2937",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          gap: 12,
        }}
      >
        <div>
          <h1 style={{ margin: 0, fontSize: 18 }}>Map &amp; Zones</h1>
          <p style={{ margin: "4px 0 0", color: "#94a3b8", fontSize: 13 }}>
            Drag to move · drag corners to resize · shift-click for multi-select
          </p>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          {/* Add zone */}
          {showAddForm ? (
            <>
              <input
                autoFocus
                value={newZoneName}
                onChange={(e) => setNewZoneName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleAddZone();
                  if (e.key === "Escape") setShowAddForm(false);
                }}
                placeholder="Zone name"
                style={{
                  padding: "4px 8px",
                  borderRadius: 4,
                  border: "1px solid #334155",
                  background: "#1e293b",
                  color: "#f1f5f9",
                  fontSize: 13,
                }}
              />
              <button
                onClick={handleAddZone}
                disabled={createZone.isPending}
                style={{
                  padding: "4px 12px",
                  borderRadius: 4,
                  border: "none",
                  cursor: "pointer",
                  fontSize: 13,
                  background: "#6366f1",
                  color: "#fff",
                  opacity: createZone.isPending ? 0.6 : 1,
                }}
              >
                {createZone.isPending ? "Saving…" : "Add"}
              </button>
              <button
                onClick={() => setShowAddForm(false)}
                style={{
                  padding: "4px 8px",
                  borderRadius: 4,
                  border: "none",
                  cursor: "pointer",
                  fontSize: 13,
                  background: "transparent",
                  color: "#94a3b8",
                }}
              >
                Cancel
              </button>
            </>
          ) : (
            <button
              data-testid="add-zone-btn"
              onClick={() => setShowAddForm(true)}
              style={{
                padding: "4px 14px",
                borderRadius: 4,
                border: "1px solid #334155",
                cursor: "pointer",
                fontSize: 13,
                background: "transparent",
                color: "#f1f5f9",
              }}
            >
              + Add zone
            </button>
          )}

          {/* View tabs */}
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
        </div>
      </header>

      <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
        <div style={{ flex: 1, position: "relative", background: "#0f172a" }}>
          <KonvaZoneCanvas
            zones={zones}
            selectedIds={selectedIds}
            onSelect={setSelectedIds}
            onChange={handleChange}
            viewMode={view}
          />
        </div>
        <ZoneDetailSidebar zone={selectedZone} />
      </div>
    </div>
  );
}
