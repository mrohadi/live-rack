import { useLocation } from "react-router-dom";
import { Icon, Icons } from "./Icon";

const PAGE_LABELS: Record<string, string> = {
  "/": "Overview",
  "/map": "Map & Zones",
  "/scanner": "Scanner",
  "/inventory": "Inventory",
  "/tasks": "Tasks",
  "/pipelines": "Pipelines",
  "/analytics": "Analytics",
  "/integrations": "Integrations",
  "/users": "Users & Access",
};

interface TopbarProps {
  density: "Compact" | "Balanced" | "Roomy";
  onDensityChange: (d: "Compact" | "Balanced" | "Roomy") => void;
  onOpenSearch: () => void;
  notifCount?: number;
  onClearNotifs?: () => void;
}

const DENSITIES = ["Compact", "Balanced", "Roomy"] as const;

export function Topbar({
  density,
  onDensityChange,
  onOpenSearch,
  notifCount = 0,
  onClearNotifs,
}: TopbarProps) {
  const { pathname } = useLocation();
  const label = PAGE_LABELS[pathname] ?? pathname.slice(1);

  return (
    <header className="topbar">
      <div className="crumbs">
        <span>Store #14</span>
        <span>›</span>
        <span className="now">{label}</span>
      </div>

      <button type="button" className="search" onClick={onOpenSearch}>
        <Icon d={Icons.search} size={14} />
        <span className="search-placeholder">Search SKUs, zones, tasks…</span>
        <kbd>⌘K</kbd>
      </button>

      <div className="topbar-actions">
        <div className="tabs">
          {DENSITIES.map((d) => (
            <div
              key={d}
              className={`tab${density === d ? " active" : ""}`}
              onClick={() => onDensityChange(d)}
            >
              {d}
            </div>
          ))}
        </div>
        <button className="icon-btn" title="Settings" type="button">
          <Icon d={Icons.cog} />
        </button>
        <button
          className={`icon-btn${notifCount > 0 ? " dot" : ""}`}
          title={notifCount > 0 ? `${notifCount} new notifications` : "Notifications"}
          type="button"
          onClick={onClearNotifs}
          style={{ position: "relative" }}
        >
          <Icon d={Icons.bell} />
          {notifCount > 0 && (
            <span
              aria-label={`${notifCount} unread notifications`}
              style={{
                position: "absolute",
                top: -2,
                right: -2,
                minWidth: 16,
                height: 16,
                padding: "0 4px",
                borderRadius: 8,
                background: "var(--primary)",
                color: "#fff",
                fontSize: 10,
                lineHeight: "16px",
                textAlign: "center",
                fontWeight: 600,
              }}
            >
              {notifCount > 9 ? "9+" : notifCount}
            </span>
          )}
        </button>
      </div>
    </header>
  );
}
