import { useEffect, useRef, useState } from "react";
import type { Zone } from "./types";

interface Props {
  zone: Zone | null;
  onDelete?: (id: string) => void;
  onOpen?: (id: string) => void;
  onAddItem?: (zoneId: string) => void;
  onEdit?: (zone: Zone) => void;
  onAssignTask?: (zone: Zone) => void;
}

/** Right-hand zone inspector. All mutating actions go through onEdit / onDelete / onAddItem / onAssignTask. */
export function ZoneDetailSidebar({
  zone,
  onDelete,
  onOpen,
  onAddItem,
  onEdit,
  onAssignTask,
}: Props) {
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    const close = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuOpen(false);
    };
    document.addEventListener("mousedown", close);
    return () => document.removeEventListener("mousedown", close);
  }, [menuOpen]);

  if (!zone) {
    return (
      <aside className="flex w-72 shrink-0 items-center justify-center border-l border-border bg-surface text-sm text-muted-foreground">
        Select a zone to inspect
      </aside>
    );
  }

  const fill =
    zone.items != null && zone.capacity ? Math.round((zone.items / zone.capacity) * 100) : 0;
  const { constraints } = zone;

  return (
    <aside className="flex w-72 shrink-0 flex-col border-l border-border bg-surface">
      {/* Header */}
      <div className="flex items-start gap-2 border-b border-border p-4">
        <div className="min-w-0 flex-1">
          {renaming ? (
            <input
              autoFocus
              value={draftName}
              onChange={(e) => setDraftName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && draftName.trim()) {
                  onRename?.(zone.id, draftName.trim());
                  setRenaming(false);
                }
                if (e.key === "Escape") setRenaming(false);
              }}
              onBlur={() => setRenaming(false)}
              className="w-full rounded-md border border-border bg-background px-2 py-1 text-sm text-foreground"
            />
          ) : (
            <div className="truncate text-sm font-semibold text-foreground">{zone.name}</div>
          )}
          <div className="mt-0.5 text-xs capitalize text-muted-foreground">
            {zone.type} · zone {zone.id}
          </div>
        </div>

        {/* ⋯ actions menu */}
        <div ref={menuRef} className="relative">
          <button
            type="button"
            aria-label="Zone actions"
            onClick={() => setMenuOpen((o) => !o)}
            className={`flex items-center gap-1 rounded-md border px-2 py-1 text-sm font-medium transition ${
              menuOpen
                ? "border-primary bg-primary/10 text-primary"
                : "border-border bg-muted text-foreground hover:border-primary hover:text-primary"
            }`}
          >
            ⋯
          </button>
          {menuOpen && (
            <div className="absolute right-0 z-20 mt-1 w-40 overflow-hidden rounded-md border border-border bg-surface shadow-md">
              <button
                type="button"
                onClick={() => {
                  onEdit?.(zone);
                  setMenuOpen(false);
                }}
                className="block w-full px-3 py-2 text-left text-sm text-foreground hover:bg-muted"
              >
                Edit zone…
              </button>
              <button
                type="button"
                onClick={() => {
                  onAddItem?.(zone.id);
                  setMenuOpen(false);
                }}
                className="block w-full px-3 py-2 text-left text-sm text-foreground hover:bg-muted"
              >
                + Add item…
              </button>
              <div className="my-1 border-t border-border" />
              <button
                type="button"
                onClick={() => {
                  onDelete?.(zone.id);
                  setMenuOpen(false);
                }}
                className="block w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-muted"
              >
                Delete zone
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Stats */}
      <div className="flex-1 overflow-y-auto p-4">
        <KV label="Capacity" value={`${zone.items ?? 0} / ${zone.capacity ?? 0}`} />
        <KV label="Fill" value={`${fill}%`} />
        {zone.sales != null && <KV label="Sales today" value={`$${zone.sales.toLocaleString()}`} />}
        {zone.dwell && <KV label="Avg. dwell" value={zone.dwell} />}
        {zone.misplaced != null && <KV label="Misplaced" value={String(zone.misplaced)} />}
        {zone.lastScan && <KV label="Last scan" value={zone.lastScan} mono />}

        <div className="mt-4">
          <div className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
            Constraints
          </div>
        </div>

        {constraints && Object.keys(constraints).length > 0 ? (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {constraints.allowedCategories?.map((cat) => (
              <Chip key={cat} accent>
                {cat} only
              </Chip>
            ))}
            {constraints.maxSKUs != null && <Chip>Max {constraints.maxSKUs} SKUs</Chip>}
            {constraints.climate && <Chip>Climate: {constraints.climate}</Chip>}
          </div>
        ) : (
          <p className="mt-2 text-xs text-muted-foreground">No constraints set.</p>
        )}
      </div>

      {/* Footer */}
      <div className="flex items-center gap-2 border-t border-border p-4">
        <button
          type="button"
          onClick={() => onOpen?.(zone.id)}
          className="flex-1 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition hover:bg-muted"
        >
          Open zone
        </button>
        <button
          type="button"
          onClick={() => onAssignTask?.(zone)}
          className="flex-1 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
        >
          Assign task
        </button>
      </div>
    </aside>
  );
}

function KV({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex items-center justify-between py-1 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className={`text-foreground ${mono ? "font-mono text-xs" : ""}`}>{value}</span>
    </div>
  );
}

function Chip({ children, accent }: { children: React.ReactNode; accent?: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs ${
        accent ? "bg-primary/15 text-primary" : "bg-muted text-muted-foreground"
      }`}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {children}
    </span>
  );
}
