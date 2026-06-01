import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

test.describe("Scanner — controls render", () => {
  test.beforeEach(async ({ page }) => {
    // Accept any /scan post without hitting the backend.
    await page.route("**/api/v1/scan", async (route) => {
      await route.fulfill({ status: 201, json: { ok: true } });
    });
    await seedOidcSession(page);
    await page.goto("/scanner");
    await page.waitForLoadState("networkidle");
  });

  test("mounts the camera + WebHID controls and toggles capture", async ({ page }) => {
    await expect(page.getByText("Camera + WebHID barcode scanning — P2")).toBeVisible();
    await expect(page.getByRole("button", { name: "Connect Zebra" })).toBeVisible();

    const toggle = page.getByRole("button", { name: "Start camera" });
    await expect(toggle).toBeVisible();
    await toggle.click();
    await expect(page.getByRole("button", { name: "Stop camera" })).toBeVisible();
  });
});
