import { useEffect, useState } from "react";
import { Outlet } from "react-router-dom";
import { CommandPalette } from "../../features/search/CommandPalette";
import { MobileTabbar } from "./MobileTabbar";
import { MobileTopbar } from "./MobileTopbar";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

type Density = "Compact" | "Balanced" | "Roomy";

export function AppShell() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [density, setDensity] = useState<Density>("Balanced");
  const [searchOpen, setSearchOpen] = useState(false);

  const closeDrawer = () => setDrawerOpen(false);

  // ⌘K / Ctrl+K toggles the command palette from anywhere.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setSearchOpen((v) => !v);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <div className="app" data-drawer={drawerOpen ? "open" : "closed"}>
      {/* Mobile drawer backdrop */}
      <div className="mobile-drawer-backdrop" onClick={closeDrawer} aria-hidden="true" />

      <Sidebar onNavigate={closeDrawer} />

      <div className="main">
        <Topbar
          density={density}
          onDensityChange={setDensity}
          onOpenSearch={() => setSearchOpen(true)}
        />
        <MobileTopbar onOpenDrawer={() => setDrawerOpen(true)} />

        <div className="content">
          <Outlet />
        </div>

        <MobileTabbar />
      </div>

      <CommandPalette open={searchOpen} onClose={() => setSearchOpen(false)} />
    </div>
  );
}
