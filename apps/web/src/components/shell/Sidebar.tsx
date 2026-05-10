import { useUser } from "@clerk/clerk-react";
import { NavLink } from "react-router-dom";
import { BrandMark } from "./BrandMark";
import { Icon, Icons } from "./Icon";

interface SidebarProps {
  accent?: string;
  onNavigate?: () => void;
}

const NAV_SECTIONS = [
  {
    label: "Operate",
    items: [
      { to: "/", name: "Overview", icon: Icons.dash, badge: null, end: true },
      { to: "/map", name: "Map & Zones", icon: Icons.map, badge: "11", end: false },
      { to: "/scanner", name: "Scanner", icon: Icons.scan, badge: null, end: false },
      { to: "/inventory", name: "Inventory", icon: Icons.box, badge: "2.5k", end: false },
    ],
  },
  {
    label: "Workflows",
    items: [
      { to: "/tasks", name: "Tasks", icon: Icons.task, badge: "6", end: false },
      { to: "/pipelines", name: "Pipelines", icon: Icons.pipe, badge: null, end: false },
    ],
  },
  {
    label: "Insights & Setup",
    items: [
      { to: "/analytics", name: "Analytics", icon: Icons.chart, badge: null, end: false },
      { to: "/integrations", name: "Integrations", icon: Icons.plug, badge: "7", end: false },
      { to: "/users", name: "Users & Access", icon: Icons.user, badge: null, end: false },
    ],
  },
] as const;

export function Sidebar({ accent = "#2563eb", onNavigate }: SidebarProps) {
  const { user } = useUser();
  const initials = user
    ? `${user.firstName?.[0] ?? ""}${user.lastName?.[0] ?? ""}`.toUpperCase()
    : "?";

  return (
    <aside className="sidebar">
      <div className="brand">
        <BrandMark accent={accent} />
        <div className="brand-name">live-rack</div>
        <div className="brand-sub">v0.1</div>
      </div>

      {NAV_SECTIONS.map((section) => (
        <div className="nav-section" key={section.label}>
          <div className="nav-label">{section.label}</div>
          {section.items.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              onClick={onNavigate}
              className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
            >
              <Icon d={item.icon} />
              <span>{item.name}</span>
              {item.badge && <span className="nav-badge">{item.badge}</span>}
            </NavLink>
          ))}
        </div>
      ))}

      <div className="sidebar-footer">
        <div className="user-chip">
          <div className="avatar">{initials || "?"}</div>
          <div style={{ minWidth: 0 }}>
            <div className="user-name">{user?.fullName ?? "—"}</div>
            <div className="user-role">{user?.primaryEmailAddress?.emailAddress ?? ""}</div>
          </div>
        </div>
      </div>
    </aside>
  );
}
