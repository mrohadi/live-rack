import { useCallback, useEffect, useState } from "react";
import { Outlet } from "react-router-dom";
import { useToast } from "../feedback/toast-context";
import { useTaskNotifications } from "../../lib/useTaskNotifications";
import type { TaskNotification } from "../../lib/ws";
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
  const [unread, setUnread] = useState(0);
  const toast = useToast();

  const closeDrawer = () => setDrawerOpen(false);

  // Live task notifications → toast + unread bell badge.
  const onNotify = useCallback(
    (n: TaskNotification) => {
      setUnread((c) => c + 1);
      const msg = n.kind === "deadline" ? `Task due soon: ${n.title}` : `Task assigned: ${n.title}`;
      toast.info(msg);
    },
    [toast],
  );
  useTaskNotifications(onNotify);

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
          notifCount={unread}
          onClearNotifs={() => setUnread(0)}
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
