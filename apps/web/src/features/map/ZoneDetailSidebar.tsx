import { useEffect, useRef, useState } from "react";
import type { Zone } from "./types";

interface Props {
  zone: Zone | null;
  onRename?: (id: string, name: string) => void;
  onDelete?: (id: string) => void;
  onOpen?: (id: string) => void;
  onAddItem?: (zoneId: string) => void;
}

/** Right-hand zone inspector with capacity, fill, sales, dwell, misplaced, last
 *  scan, constraint chips, and Open / Add item / Assign task actions. */
export function ZoneDetailSidebar({ zone, onRename, onDelete, onOpen, onAddItem }: Props) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [renaming, setRenaming] = useState(false);
  const [draftName, setDraftName] = useState("");
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
        <div ref={menuRef} className="relative">
          <button
            type="button"
            aria-label="more"
            onClick={() => setMenuOpen((o) => !o)}
            className="rounded-md p-1 text-muted-foreground transition hover:bg-muted hover:text-foreground"
          >
            ⋯
          </button>
          {menuOpen && (
            <div className="absolute right-0 z-20 mt-1 w-32 overflow-hidden rounded-md border border-border bg-surface shadow-md">
              <button
                type="button"
                onClick={() => {
                  setDraftName(zone.name);
                  setRenaming(true);
                  setMenuOpen(false);
                }}
                className="block w-full px-3 py-2 text-left text-sm text-foreground hover:bg-muted"
              >
                Rename
              </button>
              <button
                type="button"
                onClick={() => {
                  onDelete?.(zone.id);
                  setMenuOpen(false);
                }}
                className="block w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-muted"
              >
                Delete
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        <KV label="Capacity" value={`${zone.items ?? 0} / ${zone.capacity ?? 0}`} />
        <KV label="Fill" value={`${fill}%`} />
        {zone.sales != null && <KV label="Sales today" value={`$${zone.sales.toLocaleString()}`} />}
        {zone.dwell && <KV label="Avg. dwell" value={zone.dwell} />}
        {zone.misplaced != null && <KV label="Misplaced" value={String(zone.misplaced)} />}
        {zone.lastScan && <KV label="Last scan" value={zone.lastScan} mono />}

        {constraints && (
          <>
            <div className="mt-4 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
              Constraints
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {constraints.allowedCategories?.map((cat) => (
                <Chip key={cat} accent>
                  {cat} only
                </Chip>
              ))}
              {constraints.maxSKUs != null && <Chip>Max {constraints.maxSKUs} SKUs</Chip>}
              {constraints.climate && <Chip>Climate: {constraints.climate}</Chip>}
            </div>
          </>
        )}
      </div>

      <div className="flex flex-col gap-2 border-t border-border p-4">
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => onOpen?.(zone.id)}
            className="text-sm font-medium text-foreground hover:underline"
          >
            Open zone
          </button>
          <button
            type="button"
            className="ml-auto rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
          >
            Assign task
          </button>
        </div>
        <button
          type="button"
          onClick={() => onAddItem?.(zone.id)}
          className="w-full rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition hover:bg-muted"
        >
          + Add item to zone
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
