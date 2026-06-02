import { expect, test } from "@playwright/test";

// No OIDC session seeded → AuthGuard renders the branded landing instead of
// redirecting straight to Zitadel.
test.describe("Welcome — signed-out landing", () => {
  test("offers sign in and routes to signup", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Sign in" })).toBeVisible();

    await page.getByRole("link", { name: "Create a workspace" }).click();
    await expect(page).toHaveURL(/\/signup$/);
    await expect(page.getByRole("heading", { name: "Create your workspace" })).toBeVisible();
  });
});
