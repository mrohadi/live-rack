import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const API = "**/api/v1/stores/*/inventory";

const ROWS = [
  {
    id: "i1",
    sku: "LR-1001",
    name: "Vintage Lamp",
    category: "lighting",
    zone_id: "aaaaaaaa-1111",
    status: "active",
    qty: 12,
    velocity: "hot",
    updated_at: new Date().toISOString(),
  },
  {
    id: "i2",
    sku: "LR-2002",
    name: "Oak Chair",
    category: "furniture",
    zone_id: "bbbbbbbb-2222",
    status: "discontinued",
    qty: 0,
    velocity: "dead",
    updated_at: new Date().toISOString(),
  },
];

test.describe("Inventory — table + filters", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(API, async (route) => {
      if (route.request().method() === "GET") {
        await route.fulfill({ json: ROWS });
      } else {
        await route.continue();
      }
    });
    await seedOidcSession(page);
    await page.goto("/inventory");
    await page.waitForLoadState("networkidle");
  });

  test("renders rows and filters by velocity", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Inventory" })).toBeVisible();
    await expect(page.getByTestId("inventory-row")).toHaveCount(2);
    await expect(page.getByText("LR-1001")).toBeVisible();
    await expect(page.getByTestId("qty-LR-1001")).toHaveText("12");

    // Filter to the dead band → only the Oak Chair row remains.
    await page.getByLabel("Velocity").selectOption("dead");
    await expect(page.getByTestId("inventory-row")).toHaveCount(1);
    await expect(page.getByText("LR-2002")).toBeVisible();
    await expect(page.getByText("LR-1001")).toBeHidden();
  });
});
