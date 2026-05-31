import { Navigate, type RouteObject } from "react-router-dom";
import { AuthGuard } from "./components/AuthGuard";
import { AppShell } from "./components/shell";

// AuthProvider processes the OIDC redirect automatically; this just shows a spinner.
function CallbackPage() {
  return (
    <div className="flex h-screen items-center justify-center">
      <span className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
    </div>
  );
}

export const routes: RouteObject[] = [
  {
    path: "/callback",
    element: <CallbackPage />,
  },
  // Zitadel issues relative login redirects; if one lands on the app origin, recover to home.
  {
    path: "/ui/*",
    element: <Navigate to="/" replace />,
  },
  {
    path: "/",
    element: <AuthGuard />,
    children: [
      {
        element: <AppShell />,
        children: [
          {
            index: true,
            lazy: () =>
              import("./features/dashboard/DashboardPage").then((m) => ({
                Component: m.DashboardPage,
              })),
          },
          {
            path: "map",
            lazy: () => import("./features/map/MapPage").then((m) => ({ Component: m.MapPage })),
          },
          {
            path: "scanner",
            lazy: () =>
              import("./features/scanner/ScannerPage").then((m) => ({
                Component: m.ScannerPage,
              })),
          },
          {
            path: "inventory",
            lazy: () =>
              import("./features/inventory/InventoryPage").then((m) => ({
                Component: m.InventoryPage,
              })),
          },
          {
            path: "tasks",
            lazy: () =>
              import("./features/tasks/TasksPage").then((m) => ({ Component: m.TasksPage })),
          },
          {
            path: "pipelines",
            lazy: () =>
              import("./features/pipelines/PipelinesPage").then((m) => ({
                Component: m.PipelinesPage,
              })),
          },
          {
            path: "analytics",
            lazy: () =>
              import("./features/analytics/AnalyticsPage").then((m) => ({
                Component: m.AnalyticsPage,
              })),
          },
          {
            path: "integrations",
            lazy: () =>
              import("./features/integrations/IntegrationsPage").then((m) => ({
                Component: m.IntegrationsPage,
              })),
          },
          {
            path: "users",
            lazy: () =>
              import("./features/users/UsersPage").then((m) => ({ Component: m.UsersPage })),
          },
        ],
      },
    ],
  },
];
