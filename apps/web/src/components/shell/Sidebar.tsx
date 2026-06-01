import { useAuth } from "react-oidc-context";
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
  const auth = useAuth();
  const profile = auth.user?.profile;
  const fullName = (profile?.name as string | undefined) ?? "";
  const email = (profile?.email as string | undefined) ?? "";
  const initials =
    fullName
      .split(" ")
      .map((p) => p[0] ?? "")
      .join("")
      .slice(0, 2)
      .toUpperCase() || "?";

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
          <div className="avatar">{initials}</div>
          <div style={{ minWidth: 0 }}>
            <div className="user-name">{fullName || "—"}</div>
            <div className="user-role">{email}</div>
          </div>
          <button
            type="button"
            aria-label="Sign out"
            title="Sign out"
            onClick={() => {
              void auth.signoutRedirect();
            }}
            className="ml-auto rounded-md p-1.5 text-muted-foreground transition hover:bg-muted hover:text-foreground"
          >
            <svg
              viewBox="0 0 24 24"
              className="h-4 w-4"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <polyline points="16 17 21 12 16 7" />
              <line x1="21" y1="12" x2="9" y2="12" />
            </svg>
          </button>
        </div>
      </div>
    </aside>
  );
}
