import { expect, test } from "@playwright/test";

// Public auth screens (custom UI replacing the Zitadel hosted pages). Backend
// calls are route-stubbed; these assert the UI flow end to end. No OIDC session.

const STRONG = "Sup3rSecret!";

test.describe("Forgot password", () => {
  test("emails a reset link and shows confirmation", async ({ page }) => {
    let posted: Record<string, unknown> | null = null;
    await page.route("**/api/v1/password/forgot", async (route) => {
      posted = route.request().postDataJSON();
      await route.fulfill({ status: 204, body: "" });
    });

    await page.goto("/forgot-password");
    await page.waitForLoadState("networkidle");
    await expect(page.getByRole("heading", { name: "Forgot password" })).toBeVisible();

    await page.getByLabel("Email").fill("Ada@Acme.test");
    await page.getByRole("button", { name: "Send reset link" }).click();

    await expect(page.getByRole("heading", { name: "Check your email" })).toBeVisible();
    expect(posted).toMatchObject({ email: "ada@acme.test" });
  });
});

test.describe("Reset password", () => {
  test("sets a new password from a reset link", async ({ page }) => {
    let posted: Record<string, unknown> | null = null;
    await page.route("**/api/v1/password/reset", async (route) => {
      posted = route.request().postDataJSON();
      await route.fulfill({ status: 204, body: "" });
    });

    await page.goto("/reset-password?code=RC123&userID=u-1");
    await page.waitForLoadState("networkidle");
    await expect(page.getByRole("heading", { name: "Reset password" })).toBeVisible();

    await page.getByLabel("New password").fill(STRONG);
    await page.getByLabel("Confirm password").fill(STRONG);
    await page.getByRole("button", { name: "Update password" }).click();

    await expect(page.getByRole("heading", { name: "Password updated" })).toBeVisible();
    expect(posted).toMatchObject({ user_id: "u-1", code: "RC123", password: STRONG });
  });

  test("rejects an incomplete link", async ({ page }) => {
    await page.goto("/reset-password?code=RC123");
    await page.waitForLoadState("networkidle");
    await expect(page.getByText("The link is missing required details.")).toBeVisible();
  });
});

test.describe("Invite onboarding", () => {
  test("sets password then enrolls an authenticator", async ({ page }) => {
    const calls: string[] = [];
    await page.route("**/api/v1/onboard/complete", async (route) => {
      calls.push("complete");
      await route.fulfill({ status: 204, body: "" });
    });
    await page.route("**/api/v1/onboard/totp/start", async (route) => {
      calls.push("start");
      await route.fulfill({
        status: 200,
        json: { uri: "otpauth://totp/live-rack:e2e?secret=ABCSECRET", secret: "ABCSECRET" },
      });
    });
    let verifyBody: Record<string, unknown> | null = null;
    await page.route("**/api/v1/onboard/totp/verify", async (route) => {
      verifyBody = route.request().postDataJSON();
      await route.fulfill({ status: 204, body: "" });
    });

    await page.goto("/verify-email?code=EV123&userID=u-9&orgID=o-1");
    await page.waitForLoadState("networkidle");
    await expect(page.getByRole("heading", { name: "Verify your email" })).toBeVisible();

    await page.getByLabel("New password").fill(STRONG);
    await page.getByLabel("Confirm password").fill(STRONG);
    await page.getByRole("button", { name: "Continue" }).click();

    // Transitions to the authenticator step.
    await expect(page.getByRole("heading", { name: "Set up authenticator" })).toBeVisible();
    await expect(page.getByText("ABCSECRET")).toBeVisible();

    await page.getByLabel("6-digit code").fill("123456");
    await page.getByRole("button", { name: "Verify & finish" }).click();

    await expect(page.getByRole("heading", { name: "You're all set" })).toBeVisible();
    expect(calls).toEqual(["complete", "start"]);
    expect(verifyBody).toMatchObject({ user_id: "u-9", code: "123456", password: STRONG });
  });
});
