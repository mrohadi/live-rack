import { useState } from "react";
import { Outlet } from "react-router-dom";
import { MobileTabbar } from "./MobileTabbar";
import { MobileTopbar } from "./MobileTopbar";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

type Density = "Compact" | "Balanced" | "Roomy";

export function AppShell() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [density, setDensity] = useState<Density>("Balanced");

  const closeDrawer = () => setDrawerOpen(false);

  return (
    <div className="app" data-drawer={drawerOpen ? "open" : "closed"}>
      {/* Mobile drawer backdrop */}
      <div
        className="mobile-drawer-backdrop"
        onClick={closeDrawer}
        aria-hidden="true"
      />

      <Sidebar onNavigate={closeDrawer} />

      <div className="main">
        <Topbar density={density} onDensityChange={setDensity} />
        <MobileTopbar onOpenDrawer={() => setDrawerOpen(true)} />

        <div className="content">
          <Outlet />
        </div>

        <MobileTabbar />
      </div>
    </div>
  );
}
