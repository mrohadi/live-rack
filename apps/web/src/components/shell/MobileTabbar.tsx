import { NavLink } from "react-router-dom";
import { Icon, Icons } from "./Icon";

const TABS = [
  { to: "/", name: "Home", icon: Icons.dash, end: true, dot: false },
  { to: "/map", name: "Map", icon: Icons.map, end: false, dot: false },
  { to: "/scanner", name: "Scan", icon: Icons.scan, end: false, dot: true },
  { to: "/tasks", name: "Tasks", icon: Icons.task, end: false, dot: false },
  { to: "/inventory", name: "Items", icon: Icons.box, end: false, dot: false },
] as const;

export function MobileTabbar() {
  return (
    <nav className="mobile-tabbar">
      {TABS.map((tab) => (
        <NavLink
          key={tab.to}
          to={tab.to}
          end={tab.end}
          className={({ isActive }) => `m-tab${isActive ? " active" : ""}`}
        >
          {({ isActive }) => (
            <>
              <Icon d={tab.icon} size={20} />
              <span>{tab.name}</span>
              {tab.dot && !isActive && <span className="m-tab-dot" />}
            </>
          )}
        </NavLink>
      ))}
    </nav>
  );
}
