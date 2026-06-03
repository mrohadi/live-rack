import { useEffect, useRef, useState } from "react";
import { useAuth } from "react-oidc-context";
import { NavLink } from "react-router-dom";
import { useStores } from "../../features/stores/useStores";
import { getSelectedStoreId, setSelectedStoreId } from "../../lib/storeState";
import { isAdmin } from "../../lib/roles";
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
      { to: "/picking", name: "Picking", icon: Icons.pick, badge: null, end: false },
      { to: "/waves", name: "Waves", icon: Icons.wave, badge: null, end: false },
      { to: "/shipments", name: "Dispatch", icon: Icons.truck, badge: null, end: false },
      { to: "/pipelines", name: "Pipelines", icon: Icons.pipe, badge: null, end: false },
    ],
  },
  {
    label: "Insights & Setup",
    items: [
      { to: "/analytics", name: "Analytics", icon: Icons.chart, badge: null, end: false },
      { to: "/integrations", name: "Integrations", icon: Icons.plug, badge: "7", end: false },
      {
        to: "/users",
        name: "Users & Access",
        icon: Icons.user,
        badge: null,
        end: false,
        adminOnly: true,
      },
      {
        to: "/stores",
        name: "Stores",
        icon: Icons.map,
        badge: null,
        end: false,
        adminOnly: true,
      },
    ],
  },
] as const;

export function Sidebar({ accent = "#2563eb", onNavigate }: SidebarProps) {
  const auth = useAuth();
  const profile = auth.user?.profile;
  const admin = isAdmin(profile);
  const { data: stores = [] } = useStores();
  const [switcherOpen, setSwitcherOpen] = useState(false);
  const switcherRef = useRef<HTMLDivElement>(null);
  const selectedId = getSelectedStoreId();
  const currentStore = stores.find((s) => s.id === selectedId) ?? stores[0];

  const fullName = (profile?.name as string | undefined) ?? "";
  const email = (profile?.email as string | undefined) ?? "";
  const initials =
    fullName
      .split(" ")
      .map((p) => p[0] ?? "")
      .join("")
      .slice(0, 2)
      .toUpperCase() || "?";

  // Close on outside click
  useEffect(() => {
    if (!switcherOpen) return;
    const handler = (e: MouseEvent) => {
      if (switcherRef.current && !switcherRef.current.contains(e.target as Node)) {
        setSwitcherOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [switcherOpen]);

  return (
    <aside className="sidebar">
      <div className="brand">
        <BrandMark accent={accent} />
        <div className="brand-name">live-rack</div>
        <div className="brand-sub">v0.1</div>
      </div>

      {/* Store switcher — absolute dropdown so it never pushes nav down */}
      <div ref={switcherRef} style={{ position: "relative", padding: "10px 12px 6px" }}>
        <button
          type="button"
          onClick={() => setSwitcherOpen((o) => !o)}
          style={{
            display: "flex",
            alignItems: "center",
            gap: 6,
            width: "100%",
            padding: "6px 8px",
            borderRadius: 6,
            border: "1px solid var(--border)",
            background: "var(--surface)",
            cursor: "pointer",
            fontSize: 12,
            color: "var(--foreground)",
            textAlign: "left",
          }}
        >
          <Icon d={Icons.map} size={13} />
          <span
            style={{
              flex: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {currentStore?.name ?? "Select store…"}
          </span>
          <svg
            viewBox="0 0 10 6"
            width={10}
            height={10}
            fill="currentColor"
            aria-hidden
            style={{ transform: switcherOpen ? "rotate(180deg)" : "none", transition: "transform .15s" }}
          >
            <path d="M0 0l5 6 5-6z" />
          </svg>
        </button>

        {switcherOpen && (
          <div
            style={{
              position: "absolute",
              top: "calc(100% - 4px)",
              left: 12,
              right: 12,
              zIndex: 40,
              background: "var(--surface, #fff)",
              border: "1px solid var(--border)",
              borderRadius: 6,
              overflow: "hidden",
              boxShadow: "0 6px 16px rgba(0,0,0,.14)",
            }}
          >
            {stores.map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => {
                  setSelectedStoreId(s.id);
                  setSwitcherOpen(false);
                  window.location.reload();
                }}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                  width: "100%",
                  padding: "8px 10px",
                  textAlign: "left",
                  fontSize: 12,
                  color:
                    s.id === (selectedId ?? stores[0]?.id) ? "var(--primary)" : "var(--foreground)",
                  background:
                    s.id === (selectedId ?? stores[0]?.id)
                      ? "var(--primary-faint, #eff6ff)"
                      : "transparent",
                  border: "none",
                  cursor: "pointer",
                }}
              >
                {s.id === (selectedId ?? stores[0]?.id) && (
                  <svg viewBox="0 0 8 8" width={8} height={8} fill="currentColor" aria-hidden>
                    <circle cx="4" cy="4" r="4" />
                  </svg>
                )}
                <span style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                  {s.name}
                </span>
              </button>
            ))}
          </div>
        )}
      </div>

      {NAV_SECTIONS.map((section) => (
        <div className="nav-section" key={section.label}>
          <div className="nav-label">{section.label}</div>
          {section.items
            .filter((item) => admin || !("adminOnly" in item && item.adminOnly))
            .map((item) => (
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
            onClick={() => void auth.signoutRedirect()}
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
