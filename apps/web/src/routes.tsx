import { SignIn, SignUp } from "@clerk/clerk-react";
import type { RouteObject } from "react-router-dom";
import { AuthGuard } from "./components/AuthGuard";

export const routes: RouteObject[] = [
  {
    path: "/sign-in",
    element: <SignIn routing="path" path="/sign-in" />,
  },
  {
    path: "/sign-up",
    element: <SignUp routing="path" path="/sign-up" />,
  },
  {
    path: "/",
    element: <AuthGuard />,
    children: [
      {
        index: true,
        lazy: () => import("./features/dashboard/DashboardPage").then((m) => ({ Component: m.DashboardPage })),
      },
      {
        path: "map",
        lazy: () => import("./features/map/MapPage").then((m) => ({ Component: m.MapPage })),
      },
      {
        path: "inventory",
        lazy: () => import("./features/inventory/InventoryPage").then((m) => ({ Component: m.InventoryPage })),
      },
      {
        path: "tasks",
        lazy: () => import("./features/tasks/TasksPage").then((m) => ({ Component: m.TasksPage })),
      },
      {
        path: "pipelines",
        lazy: () => import("./features/pipelines/PipelinesPage").then((m) => ({ Component: m.PipelinesPage })),
      },
      {
        path: "analytics",
        lazy: () => import("./features/analytics/AnalyticsPage").then((m) => ({ Component: m.AnalyticsPage })),
      },
      {
        path: "integrations",
        lazy: () => import("./features/integrations/IntegrationsPage").then((m) => ({ Component: m.IntegrationsPage })),
      },
      {
        path: "users",
        lazy: () => import("./features/users/UsersPage").then((m) => ({ Component: m.UsersPage })),
      },
    ],
  },
];
