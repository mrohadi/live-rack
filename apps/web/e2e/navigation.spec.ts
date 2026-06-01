import { expect, test, type Page } from "@playwright/test";
import { seedOidcSession } from "./auth";

// Broad GET stubs so every feature route renders without a live backend.
async function stubAll(page: Page) {
  const json = (body: unknown) => async (route: import("@playwright/test").Route) => {
    if (route.request().method() === "GET") await route.fulfill({ json: body });
    else await route.continue();
  };
  await page.route(
    "**/api/v1/sales/summary",
    json({ revenue_cents: 0, units: 0, orders: 0, spark: [] }),
  );
  await page.route("**/api/v1/stores/*/zones", json([]));
  await page.route("**/api/v1/stores/*/inventory", json([]));
  await page.route("**/api/v1/stores/*/tasks", json([]));
  await page.route("**/api/v1/stores/*/pipelines", json([]));
  await page.route("**/api/v1/analytics/heatmap*", json({ grid: [], max: 0 }));
  await page.route("**/api/v1/analytics/zones", json({ zones: [] }));
  await page.route("**/api/v1/integrations/webhooks", json([]));
  await page.route("**/api/v1/integrations", json([]));
  await page.route("**/api/v1/users", json([]));
  await page.route(
    "**/api/v1/me",
    json({
      user_id: "u1",
      role: "admin",
      mfa_verified: false,
      permissions: [],
      store_scoped: false,
      zone_scoped: false,
    }),
  );
}

// Regression sweep: every sidebar destination must route without crashing.
const ROUTES: { link: string; path: string }[] = [
  { link: "Map & Zones", path: "/map" },
  { link: "Scanner", path: "/scanner" },
  { link: "Inventory", path: "/inventory" },
  { link: "Tasks", path: "/tasks" },
  { link: "Pipelines", path: "/pipelines" },
  { link: "Analytics", path: "/analytics" },
  { link: "Integrations", path: "/integrations" },
  { link: "Users & Access", path: "/users" },
];

test.describe("Navigation — full feature sweep", () => {
  test.beforeEach(async ({ page }) => {
    await stubAll(page);
    await seedOidcSession(page);
    await page.goto("/");
    await page.waitForLoadState("networkidle");
  });

  test("dashboard loads as the index route", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Overview" })).toBeVisible();
  });

  test("navigates to every feature via the sidebar", async ({ page }) => {
    for (const { link, path } of ROUTES) {
      await page.getByRole("link", { name: link }).click();
      await expect(page).toHaveURL(new RegExp(`${path}$`));
      // No React error boundary / blank crash — the app shell stays mounted.
      await expect(page.locator(".sidebar")).toBeVisible();
    }
  });
});
