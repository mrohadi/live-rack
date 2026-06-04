import { useEffect, useRef, useState } from "react";
import { useAuth } from "react-oidc-context";
import { NavLink } from "react-router-dom";
import { useStores } from "../../features/stores/useStores";
import { getSelectedStoreId, setSelectedStoreId } from "../../lib/storeState";
import { useSidebarCounts } from "../../lib/useSidebarCounts";
import { isAdmin } from "../../lib/roles";
import { BrandMark } from "./BrandMark";
import { Icon, Icons } from "./Icon";

interface SidebarProps {
  accent?: string;
  onNavigate?: () => void;
}

export function Sidebar({ accent = "#2563eb", onNavigate }: SidebarProps) {
  const auth = useAuth();
  const profile = auth.user?.profile;
  const admin = isAdmin(profile);
  const { data: stores = [] } = useStores();
  const counts = useSidebarCounts();
  const [switcherOpen, setSwitcherOpen] = useState(false);
  const [accountOpen, setAccountOpen] = useState(false);
  const switcherRef = useRef<HTMLDivElement>(null);
  const accountRef = useRef<HTMLDivElement>(null);
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

  // Nav sections built with live counts
  const NAV_SECTIONS = [
    {
      label: "Operate",
      items: [
        { to: "/", name: "Overview", icon: Icons.dash, badge: null as string | null, end: true },
        { to: "/map", name: "Map & Zones", icon: Icons.map, badge: counts.zones, end: false },
        { to: "/scanner", name: "Scanner", icon: Icons.scan, badge: null as string | null, end: false },
        { to: "/inventory", name: "Inventory", icon: Icons.box, badge: counts.inventory, end: false },
      ],
    },
    {
      label: "Workflows",
      items: [
        { to: "/tasks", name: "Tasks", icon: Icons.task, badge: counts.tasks, end: false },
        { to: "/picking", name: "Picking", icon: Icons.pick, badge: null as string | null, end: false },
        { to: "/waves", name: "Waves", icon: Icons.wave, badge: null as string | null, end: false },
        { to: "/shipments", name: "Dispatch", icon: Icons.truck, badge: null as string | null, end: false },
        { to: "/pipelines", name: "Pipelines", icon: Icons.pipe, badge: null as string | null, end: false },
      ],
    },
    {
      label: "Insights & Setup",
      items: [
        { to: "/analytics", name: "Analytics", icon: Icons.chart, badge: null as string | null, end: false },
        { to: "/integrations", name: "Integrations", icon: Icons.plug, badge: counts.integrations, end: false },
        {
          to: "/users",
          name: "Users & Access",
          icon: Icons.user,
          badge: null as string | null,
          end: false,
          adminOnly: true,
        },
        {
          to: "/stores",
          name: "Stores",
          icon: Icons.map,
          badge: null as string | null,
          end: false,
          adminOnly: true,
        },
      ],
    },
  ];

  // Close switcher on outside click
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

  // Close account popover on outside click
  useEffect(() => {
    if (!accountOpen) return;
    const handler = (e: MouseEvent) => {
      if (accountRef.current && !accountRef.current.contains(e.target as Node)) {
        setAccountOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [accountOpen]);

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

      {/* User chip — truncated email + info popover */}
      <div className="sidebar-footer">
        <div ref={accountRef} style={{ position: "relative" }}>
          <div className="user-chip">
            <div className="avatar">{initials}</div>
            <div style={{ minWidth: 0, flex: 1 }}>
              <div
                className="user-name"
                style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
              >
                {fullName || "—"}
              </div>
              <div
                className="user-role"
                style={{
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
                title={email}
              >
                {email}
              </div>
            </div>
            {/* Info / account details */}
            <button
              type="button"
              aria-label="Account details"
              title="Account details"
              onClick={() => setAccountOpen((o) => !o)}
              className="rounded-md p-1.5 text-muted-foreground transition hover:bg-muted hover:text-foreground"
            >
              <svg
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden
              >
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="16" x2="12" y2="12" />
                <line x1="12" y1="8" x2="12.01" y2="8" />
              </svg>
            </button>
            {/* Sign out */}
            <button
              type="button"
              aria-label="Sign out"
              title="Sign out"
              onClick={() => void auth.signoutRedirect()}
              className="rounded-md p-1.5 text-muted-foreground transition hover:bg-muted hover:text-foreground"
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

          {/* Account details popover */}
          {accountOpen && (
            <div
              style={{
                position: "absolute",
                bottom: "calc(100% + 6px)",
                left: 0,
                right: 0,
                zIndex: 50,
                background: "var(--surface, #fff)",
                border: "1px solid var(--border)",
                borderRadius: 8,
                boxShadow: "0 8px 24px rgba(0,0,0,.16)",
                padding: "14px",
                fontSize: 12,
              }}
            >
              {/* Header */}
              <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 12 }}>
                <div
                  style={{
                    width: 36,
                    height: 36,
                    borderRadius: "50%",
                    background: "linear-gradient(135deg,#2563eb,#7c3aed)",
                    color: "#fff",
                    display: "grid",
                    placeItems: "center",
                    fontWeight: 700,
                    fontSize: 13,
                    flexShrink: 0,
                  }}
                >
                  {initials}
                </div>
                <div style={{ minWidth: 0 }}>
                  <div
                    style={{
                      fontWeight: 600,
                      color: "var(--foreground)",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {fullName || "—"}
                  </div>
                  <div
                    style={{
                      color: "var(--text-3)",
                      fontSize: 11,
                      wordBreak: "break-all",
                      marginTop: 1,
                    }}
                  >
                    {email}
                  </div>
                </div>
              </div>
              {/* Meta rows */}
              <div
                style={{
                  borderTop: "1px solid var(--border)",
                  paddingTop: 10,
                  display: "flex",
                  flexDirection: "column",
                  gap: 6,
                }}
              >
                {[
                  { label: "Role", value: admin ? "Admin" : "Member" },
                  { label: "Active store", value: currentStore?.name ?? "—" },
                ].map(({ label, value }) => (
                  <div key={label} style={{ display: "flex", justifyContent: "space-between", gap: 8 }}>
                    <span style={{ color: "var(--text-3)" }}>{label}</span>
                    <span style={{ color: "var(--foreground)", fontWeight: 500 }}>{value}</span>
                  </div>
                ))}
              </div>
              {/* Sign out */}
              <button
                type="button"
                onClick={() => void auth.signoutRedirect()}
                style={{
                  marginTop: 12,
                  width: "100%",
                  padding: "7px 0",
                  borderRadius: 6,
                  border: "1px solid var(--border)",
                  background: "transparent",
                  color: "var(--destructive, #ef4444)",
                  fontSize: 12,
                  fontWeight: 500,
                  cursor: "pointer",
                }}
              >
                Sign out
              </button>
            </div>
          )}
        </div>
      </div>
    </aside>
  );
}
