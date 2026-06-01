import { expect, test } from "@playwright/test";

// Custom sign-in UI backed by our login proxy (Zitadel Session API). Endpoints
// are route-stubbed; finalize redirects to a benign callback. No OIDC session.

test.describe("Login — custom sign-in", () => {
  test("password-only sign-in finalizes and redirects", async ({ page }) => {
    let finalizeBody: Record<string, unknown> | null = null;

    await page.route("**/api/v1/login/start", async (route) => {
      await route.fulfill({
        json: { session_id: "s1", session_token: "t1", mfa_required: false },
      });
    });
    await page.route("**/api/v1/login/password", async (route) => {
      await route.fulfill({ json: { session_id: "s1", session_token: "t2" } });
    });
    await page.route("**/api/v1/login/finalize", async (route) => {
      finalizeBody = route.request().postDataJSON();
      // Redirect back to the app root (no session → Welcome) to avoid a real
      // OIDC token exchange.
      await route.fulfill({ json: { callback_url: "/" } });
    });

    await page.goto("/login?authRequest=req-1");
    await page.waitForLoadState("networkidle");
    await expect(page.getByRole("heading", { name: "Sign in" })).toBeVisible();

    await page.getByLabel("Email").fill("ada@acme.test");
    await page.getByLabel("Password").fill("Sup3rSecret!");
    await page.getByRole("button", { name: "Sign in" }).click();

    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    expect(finalizeBody).toMatchObject({
      auth_request_id: "req-1",
      session_id: "s1",
      session_token: "t2",
    });
  });

  test("prompts for a TOTP code when MFA is required", async ({ page }) => {
    let totpBody: Record<string, unknown> | null = null;

    await page.route("**/api/v1/login/start", async (route) => {
      await route.fulfill({
        json: { session_id: "s1", session_token: "t1", mfa_required: true },
      });
    });
    await page.route("**/api/v1/login/password", async (route) => {
      await route.fulfill({ json: { session_id: "s1", session_token: "t2" } });
    });
    await page.route("**/api/v1/login/totp", async (route) => {
      totpBody = route.request().postDataJSON();
      await route.fulfill({ json: { session_id: "s1", session_token: "t3" } });
    });
    await page.route("**/api/v1/login/finalize", async (route) => {
      await route.fulfill({ json: { callback_url: "/" } });
    });

    await page.goto("/login?authRequest=req-2");
    await page.waitForLoadState("networkidle");

    await page.getByLabel("Email").fill("ada@acme.test");
    await page.getByLabel("Password").fill("Sup3rSecret!");
    await page.getByRole("button", { name: "Sign in" }).click();

    // MFA step appears; password did not finalize yet.
    const code = page.getByLabel("Authentication code");
    await expect(code).toBeVisible();
    await code.fill("123456");
    await page.getByRole("button", { name: "Verify" }).click();

    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    expect(totpBody).toMatchObject({ session_id: "s1", session_token: "t2", code: "123456" });
  });
});
