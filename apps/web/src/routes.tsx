import { Navigate, type RouteObject } from "react-router-dom";
import { AuthGuard } from "./components/AuthGuard";
import { CallbackPage } from "./components/CallbackPage";
import { RequireAdmin } from "./components/RequireAdmin";
import { AppShell } from "./components/shell";

export const routes: RouteObject[] = [
  {
    path: "/callback",
    element: <CallbackPage />,
  },
  {
    path: "/login",
    lazy: () => import("./features/login/LoginPage").then((m) => ({ Component: m.LoginPage })),
  },
  {
    path: "/signup",
    lazy: () => import("./features/signup/SignupPage").then((m) => ({ Component: m.SignupPage })),
  },
  {
    path: "/verify-email",
    lazy: () =>
      import("./features/onboarding/VerifyEmailPage").then((m) => ({
        Component: m.VerifyEmailPage,
      })),
  },
  {
    path: "/forgot-password",
    lazy: () =>
      import("./features/password/ForgotPasswordPage").then((m) => ({
        Component: m.ForgotPasswordPage,
      })),
  },
  {
    path: "/reset-password",
    lazy: () =>
      import("./features/password/ResetPasswordPage").then((m) => ({
        Component: m.ResetPasswordPage,
      })),
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
            path: "picking",
            lazy: () =>
              import("./features/picking/PickingPage").then((m) => ({
                Component: m.PickingPage,
              })),
          },
          {
            path: "waves",
            lazy: () =>
              import("./features/waves/WavesPage").then((m) => ({
                Component: m.WavesPage,
              })),
          },
          {
            path: "shipments",
            lazy: () =>
              import("./features/shipments/ShipmentsPage").then((m) => ({
                Component: m.ShipmentsPage,
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
            element: <RequireAdmin />,
            children: [
              {
                path: "users",
                lazy: () =>
                  import("./features/users/UsersPage").then((m) => ({ Component: m.UsersPage })),
              },
              {
                path: "stores",
                lazy: () =>
                  import("./features/stores/StoresPage").then((m) => ({
                    Component: m.StoresPage,
                  })),
              },
            ],
          },
        ],
      },
    ],
  },
];
