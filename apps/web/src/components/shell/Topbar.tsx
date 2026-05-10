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
}

const DENSITIES = ["Compact", "Balanced", "Roomy"] as const;

export function Topbar({ density, onDensityChange }: TopbarProps) {
  const { pathname } = useLocation();
  const label = PAGE_LABELS[pathname] ?? pathname.slice(1);

  return (
    <header className="topbar">
      <div className="crumbs">
        <span>Store #14</span>
        <span>›</span>
        <span className="now">{label}</span>
      </div>

      <label className="search">
        <Icon d={Icons.search} size={14} />
        <input placeholder="Search SKUs, zones, tasks…" />
        <kbd>⌘K</kbd>
      </label>

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
        <button className="icon-btn dot" title="Notifications" type="button">
          <Icon d={Icons.bell} />
        </button>
      </div>
    </header>
  );
}
