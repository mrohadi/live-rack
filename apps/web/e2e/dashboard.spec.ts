import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const API = "/api/v1/sales/summary";

const SUMMARY = {
  revenue_cents: 1234567,
  units: 842,
  orders: 311,
  spark: [3, 7, 5, 9, 4, 8, 6],
};

test.describe("Dashboard — sales overview", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(`**${API}`, async (route) => {
      if (route.request().method() === "GET") {
        await route.fulfill({ json: SUMMARY });
      } else {
        await route.continue();
      }
    });
    await seedOidcSession(page);
    await page.goto("/");
    await page.waitForLoadState("networkidle");
  });

  test("renders stat cards and the 7-day sparkline", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Overview" })).toBeVisible();

    // Formatted revenue + raw counts surface from the stubbed summary.
    await expect(page.getByText("$12,345.67")).toBeVisible();
    await expect(page.getByText("842", { exact: true })).toBeVisible();
    await expect(page.getByText("311", { exact: true })).toBeVisible();

    const spark = page.getByRole("img", { name: "7-day revenue sparkline" });
    await expect(spark).toBeVisible();
    await expect(spark.locator("polyline")).toHaveAttribute("points", /\d/);
  });
});
