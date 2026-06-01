import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const USERS = "**/api/v1/users";
const ME = "**/api/v1/me";
const INVITE = "**/api/v1/users/invite";

const ADMIN_CAPS = {
  user_id: "u1",
  role: "admin",
  mfa_verified: true,
  permissions: ["edit_users"],
  store_scoped: false,
  zone_scoped: false,
};

test.describe("Users — invite flow", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(ME, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: ADMIN_CAPS });
      else await route.continue();
    });
    await page.route(USERS, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: [] });
      else await route.continue();
    });
    await page.route("**/api/v1/users/stats", async (route) => {
      await route.fulfill({
        json: { members: 0, roles: 5, active_now: 0, pending_invites: 0, twofa_coverage: 0 },
      });
    });
    await page.route("**/api/v1/me/2fa", async (route) => {
      await route.fulfill({ status: 204, body: "" });
    });
    await page.route("**/api/v1/audit*", async (route) => {
      await route.fulfill({ json: [] });
    });

    await seedOidcSession(page);
    await page.goto("/users");
    await page.waitForLoadState("networkidle");
  });

  test("admin invites a teammate via the modal", async ({ page }) => {
    let posted: Record<string, unknown> | null = null;
    await page.route(INVITE, async (route) => {
      posted = route.request().postDataJSON();
      await route.fulfill({
        status: 201,
        json: { user_id: "zid-9", email: "new@acme.test", role: "manager", status: "invited" },
      });
    });

    await page.getByRole("button", { name: "Add user" }).click();
    await expect(page.getByRole("dialog", { name: "Invite user" })).toBeVisible();

    await page.getByLabel("Email").fill("new@acme.test");
    await page.getByLabel("Display name").fill("New Person");
    await page.getByLabel("Role").selectOption("manager");
    await page.getByRole("button", { name: "Send invite" }).click();

    // Success confirmation, then dismiss.
    await expect(page.getByText("Invitation sent")).toBeVisible();
    expect(posted).toMatchObject({ email: "new@acme.test", role: "manager" });
    await page.getByRole("button", { name: "Done" }).click();
    await expect(page.getByRole("dialog", { name: "Invite user" })).toBeHidden();
  });

  test("hides invite for non-admins", async ({ page }) => {
    await page.route(ME, async (route) => {
      await route.fulfill({ json: { ...ADMIN_CAPS, role: "staff", mfa_verified: false } });
    });
    await page.reload();
    await page.waitForLoadState("networkidle");
    await expect(page.getByRole("button", { name: "Add user" })).toBeHidden();
  });
});
