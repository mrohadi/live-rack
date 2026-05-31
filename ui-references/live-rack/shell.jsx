/* global React */
const { useState, useEffect, useMemo, useRef } = React;

// Icon factory
const Icon = ({ d, size = 16, stroke = 1.6 }) =>
<svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" strokeWidth={stroke} strokeLinecap="round" strokeLinejoin="round">
    {typeof d === 'string' ? <path d={d} /> : d}
  </svg>;

const I = {
  dash: 'M3 13l9-9 9 9M5 11v9h5v-6h4v6h5v-9',
  map: <><path d="M9 4l-6 2v14l6-2 6 2 6-2V4l-6 2-6-2z" /><path d="M9 4v14M15 6v14" /></>,
  scan: <><path d="M4 7V4h3M20 7V4h-3M4 17v3h3M20 17v3h-3" /><path d="M4 12h16" /></>,
  box: <><path d="M21 8l-9-5-9 5 9 5 9-5z" /><path d="M3 8v8l9 5 9-5V8" /><path d="M12 13v8" /></>,
  pipe: <><path d="M3 6h6v6H3zM15 6h6v6h-6zM9 18h6" /><path d="M9 9h6M12 12v6" /></>,
  task: <><path d="M9 11l3 3L22 4" /><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" /></>,
  chart: <><path d="M3 3v18h18" /><path d="M7 15l4-4 3 3 5-7" /></>,
  plug: <><path d="M9 2v4M15 2v4M7 6h10v6a5 5 0 01-10 0V6zM12 17v5" /></>,
  bell: <><path d="M6 8a6 6 0 1112 0c0 7 3 8 3 8H3s3-1 3-8" /><path d="M10 21a2 2 0 004 0" /></>,
  search: <><circle cx="11" cy="11" r="7" /><path d="M21 21l-4.3-4.3" /></>,
  plus: 'M12 5v14M5 12h14',
  filter: 'M3 5h18M6 12h12M10 19h4',
  more: <><circle cx="5" cy="12" r="1" /><circle cx="12" cy="12" r="1" /><circle cx="19" cy="12" r="1" /></>,
  arrowUp: 'M7 17L17 7M17 7H8M17 7v9',
  arrowDn: 'M17 7L7 17M7 17h9M7 17V8',
  check: 'M5 12l5 5L20 7',
  flash: 'M13 2L4 14h7l-1 8 9-12h-7l1-8z',
  clock: <><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></>,
  user: <><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0116 0" /></>,
  warn: <><path d="M12 3l10 18H2L12 3z" /><path d="M12 10v5M12 18v.01" /></>,
  cog: <><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.7 1.7 0 00.3 1.8l.1.1a2 2 0 11-2.8 2.8l-.1-.1a1.7 1.7 0 00-1.8-.3 1.7 1.7 0 00-1 1.5V21a2 2 0 11-4 0v-.1a1.7 1.7 0 00-1-1.5 1.7 1.7 0 00-1.8.3l-.1.1a2 2 0 11-2.8-2.8l.1-.1a1.7 1.7 0 00.3-1.8 1.7 1.7 0 00-1.5-1H3a2 2 0 110-4h.1a1.7 1.7 0 001.5-1 1.7 1.7 0 00-.3-1.8l-.1-.1a2 2 0 112.8-2.8l.1.1a1.7 1.7 0 001.8.3H9a1.7 1.7 0 001-1.5V3a2 2 0 114 0v.1a1.7 1.7 0 001 1.5 1.7 1.7 0 001.8-.3l.1-.1a2 2 0 112.8 2.8l-.1.1a1.7 1.7 0 00-.3 1.8V9a1.7 1.7 0 001.5 1H21a2 2 0 110 4h-.1a1.7 1.7 0 00-1.5 1z" /></>
};

// Sidebar
function Sidebar({ page, setPage, role, accent }) {
  const sections = [
  { label: 'Operate', items: [
    { id: 'dashboard', name: 'Overview', icon: I.dash },
    { id: 'map', name: 'Map & Zones', icon: I.map, badge: '11' },
    { id: 'scanner', name: 'Scanner', icon: I.scan },
    { id: 'inventory', name: 'Inventory', icon: I.box, badge: '2.5k' }]
  },
  { label: 'Workflows', items: [
    { id: 'tasks', name: 'Tasks', icon: I.task, badge: '6' },
    { id: 'pipelines', name: 'Pipelines', icon: I.pipe }]
  },
  { label: 'Insights & Setup', items: [
    { id: 'analytics', name: 'Analytics', icon: I.chart },
    { id: 'integrations', name: 'Integrations', icon: I.plug, badge: '7' },
    { id: 'users', name: 'Users & Access', icon: I.user, badge: '12' }]
  }];

  return (
    <aside className="sidebar">
      <div className="brand">
        <BrandMark accent={accent} />
        <div className="brand-name">live-rack</div>
        <div className="brand-sub">v0.4</div>
      </div>
      {sections.map((s, si) =>
      <div className="nav-section" key={si}>
          <div className="nav-label">{s.label}</div>
          {s.items.map((it) =>
        <div key={it.id}
        className={`nav-item ${page === it.id ? 'active' : ''}`}
        onClick={() => setPage(it.id)}>
              <Icon d={it.icon} />
              <span>{it.name}</span>
              {it.badge && <span className="nav-badge">{it.badge}</span>}
            </div>
        )}
        </div>
      )}
      <div className="sidebar-footer">
        <div className="user-chip">
          <div className="avatar">{role.initials}</div>
          <div style={{ minWidth: 0 }}>
            <div className="user-name">{role.name}</div>
            <div className="user-role">{role.title} · Store #14</div>
          </div>
        </div>
      </div>
    </aside>);

}

function BrandMark({ accent = '#2563eb' }) {
  return (
    <svg className="brand-mark" viewBox="0 0 32 32">
      <rect x="2" y="2" width="28" height="28" rx="7" fill={accent} />
      <rect x="6" y="9" width="20" height="3.2" rx="1" fill="white" opacity="0.95" />
      <rect x="6" y="14.4" width="13" height="3.2" rx="1" fill="white" opacity="0.7" />
      <rect x="6" y="19.8" width="20" height="3.2" rx="1" fill="white" opacity="0.95" />
      <circle cx="22" cy="16" r="1.6" fill="white" />
    </svg>);

}

// Topbar
function Topbar({ page, role, setRole, density, setDensity }) {
  const roles = [
  { id: 'manager', name: 'Avery Chen', title: 'Ops Manager', initials: 'AC' },
  { id: 'floor', name: 'Maya Reyes', title: 'Floor Lead', initials: 'MR' },
  { id: 'analyst', name: 'Theo Kim', title: 'Analyst', initials: 'TK' }];

  const labels = {
    dashboard: 'Overview', map: 'Map & Zones', scanner: 'Scanner',
    inventory: 'Inventory', tasks: 'Tasks', pipelines: 'Pipelines',
    analytics: 'Analytics', integrations: 'Integrations', users: 'Users & Access'
  };
  return (
    <header className="topbar">
      <div className="crumbs">
        <span>Store #14</span>
        <span>›</span>
        <span className="now">{labels[page]}</span>
      </div>
      <label className="search">
        <Icon d={I.search} size={14} />
        <input placeholder="Search SKUs, zones, tasks…" />
        <kbd>⌘K</kbd>
      </label>
      <div className="topbar-actions">
        <div className="tabs">
          {['Compact', 'Balanced', 'Roomy'].map((d) =>
          <div key={d} className={`tab ${density === d ? 'active' : ''}`} onClick={() => setDensity(d)}>{d}</div>
          )}
        </div>
        <button className="icon-btn" title="Settings"><Icon d={I.cog} /></button>
        <button className="icon-btn dot" title="Notifications"><Icon d={I.bell} /></button>
        <div className="role-pill">
          <span className="swatch" />
          <select value={role.id} onChange={(e) => setRole(roles.find((r) => r.id === e.target.value))}>
            {roles.map((r) => <option key={r.id} value={r.id}>{r.name} · {r.title}</option>)}
          </select>
        </div>
      </div>
    </header>);

}

// Mobile topbar
function MobileTopbar({ page, openDrawer }) {
  const labels = {
    dashboard: 'Overview', map: 'Map & Zones', scanner: 'Scanner',
    inventory: 'Inventory', tasks: 'Tasks', pipelines: 'Pipelines',
    analytics: 'Analytics', integrations: 'Integrations', users: 'Users'
  };
  return (
    <header className="mobile-topbar">
      <button className="icon-btn" onClick={openDrawer} aria-label="Menu">
        <Icon d={<><path d="M3 6h18M3 12h18M3 18h18"/></>} />
      </button>
      <div style={{flex:1, minWidth:0}}>
        <div className="m-title" style={{lineHeight:1.1}}>{labels[page]}</div>
        <div className="m-sub">Store #14 · live</div>
      </div>
      <button className="icon-btn"><Icon d={I.search} /></button>
      <button className="icon-btn dot"><Icon d={I.bell} /></button>
    </header>
  );
}

// Bottom tab bar
function MobileTabbar({ page, setPage }) {
  const tabs = [
    { id: 'dashboard', name: 'Home',     icon: I.dash },
    { id: 'map',       name: 'Map',      icon: I.map },
    { id: 'scanner',   name: 'Scan',     icon: I.scan, dot: true },
    { id: 'tasks',     name: 'Tasks',    icon: I.task },
    { id: 'inventory', name: 'Items',    icon: I.box },
  ];
  return (
    <nav className="mobile-tabbar">
      {tabs.map(t => (
        <button key={t.id} className={`m-tab ${page===t.id?'active':''}`} onClick={()=>setPage(t.id)}>
          <Icon d={t.icon} size={20} />
          <span>{t.name}</span>
          {t.dot && page!==t.id && <span className="m-tab-dot" />}
        </button>
      ))}
    </nav>
  );
}

Object.assign(window, { Icon, I, Sidebar, Topbar, BrandMark, MobileTopbar, MobileTabbar });