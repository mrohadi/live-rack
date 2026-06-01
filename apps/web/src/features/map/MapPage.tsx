import { useCallback, useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { ViewMode, Zone } from "./types";
import { useCurrentStore } from "./useCurrentStore";
import { useCreateZone, useZones, zoneKeys } from "./useZones";
import { ZoneDetailSidebar } from "./ZoneDetailSidebar";
import { ZoneMapView } from "./ZoneMapView";
import { useInventory } from "../inventory/useInventory";
import { useScanStream } from "../../lib/useScanStream";
import type { ScanRecorded } from "../../lib/ws";

const TABS: { label: string; value: ViewMode }[] = [
  { label: "Zones", value: "zones" },
  { label: "Heat", value: "heat" },
  { label: "Items", value: "items" },
];

export function MapPage() {
  const storeId = useCurrentStore();
  const { data: zones = [], isLoading } = useZones(storeId);
  const { data: items = [] } = useInventory(storeId);
  const createZone = useCreateZone(storeId);

  const qc = useQueryClient();
  const onScan = useCallback(
    (ev: ScanRecorded) => {
      qc.setQueryData<Zone[]>(zoneKeys.list(storeId), (prev) =>
        prev?.map((z) =>
          z.id === ev.zone_id
            ? {
                ...z,
                lastScan: ev.ts,
                items: ev.valid && ev.action === "place" ? (z.items ?? 0) + 1 : z.items,
                misplaced: ev.valid ? z.misplaced : (z.misplaced ?? 0) + 1,
              }
            : z,
        ),
      );
    },
    [qc, storeId],
  );
  useScanStream(onScan);

  const [view, setView] = useState<ViewMode>("zones");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [showAddForm, setShowAddForm] = useState(false);
  const [newZoneName, setNewZoneName] = useState("");

  // Default to the first zone once loaded so the detail panel is never empty.
  useEffect(() => {
    if (!selectedId && zones.length > 0) setSelectedId(zones[0].id);
  }, [zones, selectedId]);

  const handleAddZone = () => {
    if (!newZoneName.trim()) return;
    createZone.mutate(
      {
        name: newZoneName.trim(),
        type: "general",
        x: 40,
        y: 40,
        width: 20,
        height: 14,
        color: "#2563eb",
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

  const selectedZone = zones.find((z) => z.id === selectedId) ?? null;
  const itemsTracked = zones.reduce((sum, z) => sum + (z.items ?? 0), 0);

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div>
          <h1 className="text-lg font-semibold text-foreground">Map &amp; Zones</h1>
          <p className="text-xs text-muted-foreground">
            Floor 1 · {zones.length} zones · {itemsTracked.toLocaleString()} items tracked · last
            sync 12s
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="inline-flex gap-1 rounded-lg border border-border bg-surface p-1">
            {TABS.map((t) => (
              <button
                key={t.value}
                type="button"
                onClick={() => setView(t.value)}
                className={`rounded-md px-3 py-1 text-sm font-medium transition ${
                  view === t.value
                    ? "bg-primary text-white"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>

          <button
            type="button"
            className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition hover:bg-muted"
          >
            Filter
          </button>

          {showAddForm ? (
            <div className="flex items-center gap-2">
              <input
                autoFocus
                value={newZoneName}
                onChange={(e) => setNewZoneName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleAddZone();
                  if (e.key === "Escape") setShowAddForm(false);
                }}
                placeholder="Zone name"
                className="w-32 rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground"
              />
              <button
                type="button"
                onClick={handleAddZone}
                disabled={createZone.isPending}
                className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
              >
                {createZone.isPending ? "Saving…" : "Add"}
              </button>
              <button
                type="button"
                onClick={() => setShowAddForm(false)}
                className="px-2 py-1.5 text-sm text-muted-foreground hover:text-foreground"
              >
                Cancel
              </button>
            </div>
          ) : (
            <button
              type="button"
              data-testid="add-zone-btn"
              onClick={() => setShowAddForm(true)}
              className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
            >
              + New zone
            </button>
          )}
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 overflow-auto">
          {isLoading ? (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              Loading zones…
            </div>
          ) : (
            <ZoneMapView
              zones={zones}
              items={items}
              view={view}
              selectedId={selectedId}
              onSelect={setSelectedId}
            />
          )}
        </div>
        <ZoneDetailSidebar zone={selectedZone} />
      </div>
    </div>
  );
}
