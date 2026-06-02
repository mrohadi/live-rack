import { useCallback, useEffect, useState } from "react";
import { Outlet } from "react-router-dom";
import { useToast } from "../feedback/toast-context";
import { useTaskNotifications } from "../../lib/useTaskNotifications";
import { useNotificationCenter } from "../../lib/useNotificationCenter";
import type { TaskNotification } from "../../lib/ws";
import { CommandPalette } from "../../features/search/CommandPalette";
import { NotificationsPanel } from "./NotificationsPanel";
import { MobileTabbar } from "./MobileTabbar";
import { MobileTopbar } from "./MobileTopbar";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

type Density = "Compact" | "Balanced" | "Roomy";

export function AppShell() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [density, setDensity] = useState<Density>("Balanced");
  const [searchOpen, setSearchOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const toast = useToast();
  const notifs = useNotificationCenter();

  const closeDrawer = () => setDrawerOpen(false);

  // Live task notifications → notification center + toast (only when addressed
  // to the current user; the center filters by assignee).
  const onNotify = useCallback(
    (n: TaskNotification) => {
      if (!notifs.push(n)) return;
      const msg = n.kind === "deadline" ? `Task due soon: ${n.title}` : `Task assigned: ${n.title}`;
      toast.info(msg);
    },
    [notifs, toast],
  );
  useTaskNotifications(onNotify);

  const toggleNotifs = () => {
    setNotifOpen((v) => {
      const next = !v;
      if (next) notifs.markAllRead();
      return next;
    });
  };

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
          notifCount={notifs.unread}
          onToggleNotifs={toggleNotifs}
          notifPanel={
            notifOpen ? (
              <NotificationsPanel items={notifs.items} onClose={() => setNotifOpen(false)} />
            ) : null
          }
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
