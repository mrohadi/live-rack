import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import type { ViewMode, Zone } from "./types";
import { useCurrentStore } from "./useCurrentStore";
import { useCreateZone, useDeleteZone, useUpdateZone, useZones, zoneKeys } from "./useZones";
import { ZoneDetailSidebar } from "./ZoneDetailSidebar";
import { ZoneMapView, type ZoneRect } from "./ZoneMapView";
import { findOpenSlot, randomZoneColor } from "./zoneMath";
import { useToast } from "../../components/feedback/toast-context";
import { useConfirm } from "../../components/feedback/confirm-context";
import { useInventory } from "../inventory/useInventory";
import { useScanStream } from "../../lib/useScanStream";
import type { ScanRecorded } from "../../lib/ws";

const TABS: { label: string; value: ViewMode }[] = [
  { label: "Zones", value: "zones" },
  { label: "Heat", value: "heat" },
  { label: "Items", value: "items" },
];

type FillBand = "all" | "low" | "mid" | "high";

const FILL_BANDS: { label: string; value: FillBand }[] = [
  { label: "All fill levels", value: "all" },
  { label: "Low (< 50%)", value: "low" },
  { label: "Mid (50–85%)", value: "mid" },
  { label: "High (> 85%)", value: "high" },
];

export function MapPage() {
  const storeId = useCurrentStore();
  const navigate = useNavigate();
  const { data: zones = [], isLoading } = useZones(storeId);
  const { data: items = [] } = useInventory(storeId);
  const createZone = useCreateZone(storeId);
  const updateZone = useUpdateZone(storeId);
  const deleteZone = useDeleteZone(storeId);
  const toast = useToast();
  const confirm = useConfirm();

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

  const [filterOpen, setFilterOpen] = useState(false);
  const [typeFilter, setTypeFilter] = useState<Zone["type"] | "all">("all");
  const [fillFilter, setFillFilter] = useState<FillBand>("all");
  const filterRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!filterOpen) return;
    const close = (e: MouseEvent) => {
      if (filterRef.current && !filterRef.current.contains(e.target as Node)) setFilterOpen(false);
    };
    document.addEventListener("mousedown", close);
    return () => document.removeEventListener("mousedown", close);
  }, [filterOpen]);

  const zoneTypes = useMemo(() => Array.from(new Set(zones.map((z) => z.type))), [zones]);

  const visibleZones = useMemo(() => {
    return zones.filter((z) => {
      if (typeFilter !== "all" && z.type !== typeFilter) return false;
      if (fillFilter !== "all") {
        const fill = z.capacity && z.items != null ? z.items / z.capacity : 0;
        if (fillFilter === "low" && fill >= 0.5) return false;
        if (fillFilter === "mid" && (fill < 0.5 || fill > 0.85)) return false;
        if (fillFilter === "high" && fill <= 0.85) return false;
      }
      return true;
    });
  }, [zones, typeFilter, fillFilter]);

  const filterActive = typeFilter !== "all" || fillFilter !== "all";

  // Default to the first zone once loaded so the detail panel is never empty.
  useEffect(() => {
    if (!selectedId && zones.length > 0) setSelectedId(zones[0].id);
  }, [zones, selectedId]);

  const handleAddZone = () => {
    if (!newZoneName.trim()) return;
    const width = 20;
    const height = 14;
    const { x, y } = findOpenSlot(zones, width, height);
    createZone.mutate(
      {
        name: newZoneName.trim(),
        type: "general",
        x,
        y,
        width,
        height,
        color: randomZoneColor(),
        capacity: 100,
      },
      {
        onSuccess: () => {
          setNewZoneName("");
          setShowAddForm(false);
          toast.success("Zone created");
        },
      },
    );
  };

  // The backend PUT requires a full zone body; merge the new rect into it.
  const handleMove = (id: string, rect: ZoneRect) => {
    const existing = zones.find((z) => z.id === id);
    if (!existing) return;
    updateZone.mutate({ ...existing, ...rect });
  };

  const handleRename = (id: string, name: string) => {
    const existing = zones.find((z) => z.id === id);
    if (!existing) return;
    updateZone.mutate({ ...existing, name }, { onSuccess: () => toast.success("Zone renamed") });
  };

  const handleDelete = async (id: string) => {
    const zone = zones.find((z) => z.id === id);
    const ok = await confirm({
      title: "Delete zone",
      message: `Delete ${zone?.name ?? "this zone"}? This cannot be undone.`,
      confirmLabel: "Delete",
      destructive: true,
    });
    if (!ok) return;
    deleteZone.mutate(id, {
      onSuccess: () => {
        setSelectedId(null);
        toast.success("Zone deleted");
      },
      onError: () => toast.error("Failed to delete zone"),
    });
  };

  const handleOpen = (id: string) => navigate(`/inventory?zone=${encodeURIComponent(id)}`);

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

          <div ref={filterRef} className="relative">
            <button
              type="button"
              onClick={() => setFilterOpen((o) => !o)}
              className={`rounded-md border px-3 py-1.5 text-sm font-medium transition hover:bg-muted ${
                filterActive ? "border-primary text-primary" : "border-border text-foreground"
              }`}
            >
              Filter{filterActive ? " ·" : ""}
            </button>
            {filterOpen && (
              <div className="absolute right-0 z-20 mt-1 w-56 rounded-md border border-border bg-surface p-3 shadow-md">
                <label className="block text-xs font-medium text-muted-foreground">Type</label>
                <select
                  value={typeFilter}
                  onChange={(e) => setTypeFilter(e.target.value as Zone["type"] | "all")}
                  className="mt-1 w-full rounded border border-border bg-background px-2 py-1 text-sm capitalize text-foreground"
                >
                  <option value="all">All types</option>
                  {zoneTypes.map((t) => (
                    <option key={t} value={t} className="capitalize">
                      {t}
                    </option>
                  ))}
                </select>

                <label className="mt-3 block text-xs font-medium text-muted-foreground">Fill</label>
                <select
                  value={fillFilter}
                  onChange={(e) => setFillFilter(e.target.value as FillBand)}
                  className="mt-1 w-full rounded border border-border bg-background px-2 py-1 text-sm text-foreground"
                >
                  {FILL_BANDS.map((b) => (
                    <option key={b.value} value={b.value}>
                      {b.label}
                    </option>
                  ))}
                </select>

                {filterActive && (
                  <button
                    type="button"
                    onClick={() => {
                      setTypeFilter("all");
                      setFillFilter("all");
                    }}
                    className="mt-3 text-xs font-medium text-primary hover:underline"
                  >
                    Clear filters
                  </button>
                )}
              </div>
            )}
          </div>

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
              zones={visibleZones}
              items={items}
              view={view}
              selectedId={selectedId}
              onSelect={setSelectedId}
              onMove={handleMove}
            />
          )}
        </div>
        <ZoneDetailSidebar
          zone={selectedZone}
          onRename={handleRename}
          onDelete={handleDelete}
          onOpen={handleOpen}
        />
      </div>
    </div>
  );
}
