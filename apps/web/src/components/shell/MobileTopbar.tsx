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
  "/users": "Users",
};

interface MobileTopbarProps {
  onOpenDrawer: () => void;
}

export function MobileTopbar({ onOpenDrawer }: MobileTopbarProps) {
  const { pathname } = useLocation();
  const label = PAGE_LABELS[pathname] ?? pathname.slice(1);

  return (
    <header className="mobile-topbar">
      <button className="icon-btn" onClick={onOpenDrawer} aria-label="Menu" type="button">
        <Icon d={Icons.menu} />
      </button>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div className="m-title">{label}</div>
        <div className="m-sub">Store #14 · live</div>
      </div>
      <button className="icon-btn" type="button" aria-label="Search">
        <Icon d={Icons.search} />
      </button>
      <button className="icon-btn dot" type="button" aria-label="Notifications">
        <Icon d={Icons.bell} />
      </button>
    </header>
  );
}
