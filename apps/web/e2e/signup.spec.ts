import { expect, test } from "@playwright/test";

const API = "**/api/v1/signup";

test.describe("Signup — public self-service registration", () => {
  test("creates a workspace and shows the verify-email confirmation", async ({ page }) => {
    let posted: Record<string, unknown> | null = null;
    await page.route(API, async (route) => {
      posted = route.request().postDataJSON();
      await route.fulfill({
        status: 201,
        json: { org_id: "org-1", user_id: "user-1", status: "pending_verification" },
      });
    });

    // Public route — no OIDC session seeded.
    await page.goto("/signup");
    await page.waitForLoadState("networkidle");

    await expect(page.getByRole("heading", { name: "Create your workspace" })).toBeVisible();

    await page.getByLabel("Company").fill("Acme Co");
    await page.getByLabel("Work email").fill("founder@acme.test");
    await page.getByLabel("Your name").fill("Ada Founder");
    await page.getByRole("button", { name: "Create workspace" }).click();

    await expect(page.getByRole("heading", { name: "Check your email" })).toBeVisible();
    expect(posted).toMatchObject({ company: "Acme Co", email: "founder@acme.test" });
  });
});
