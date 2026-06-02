import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const HEATMAP = "**/api/v1/analytics/heatmap*";
const ZONES = "**/api/v1/analytics/zones";

const grid = Array.from({ length: 7 }, (_, d) => Array.from({ length: 24 }, (_, h) => (d + h) % 5));

const ZONES_DATA = {
  zones: [
    { zone_id: "zone-aaaa", scans: 120, picks: 90, invalid: 4, spark: [1, 2, 3, 4] },
    { zone_id: "zone-bbbb", scans: 60, picks: 40, invalid: 2, spark: [4, 3, 2, 1] },
  ],
};

test.describe("Analytics — heatmap + zone bars + signals", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(HEATMAP, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: { grid, max: 4 } });
      else await route.continue();
    });
    await page.route(ZONES, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: ZONES_DATA });
      else await route.continue();
    });

    await seedOidcSession(page);
    await page.goto("/analytics");
    await page.waitForLoadState("networkidle");
  });

  test("renders zone performance, heatmap, and the signal panel", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Analytics" })).toBeVisible();

    // Zone bars resolve from the stubbed zones response.
    await expect(page.getByTestId("zone-perf")).toBeVisible();
    await expect(page.getByText("120")).toBeVisible();
    await expect(page.getByRole("img", { name: "zone-aaaa activity" })).toBeVisible();

    // Loading placeholders are gone once data lands.
    await expect(page.getByText("Loading zones…")).toBeHidden();
    await expect(page.getByText("Loading heatmap…")).toBeHidden();

    // Signal panel renders its empty state without live WS recommendations.
    await expect(page.getByText("No active signals.")).toBeVisible();
  });
});
