import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const CAPS = {
  user_id: "u1",
  role: "admin",
  mfa_verified: true,
  permissions: ["users.edit"],
  store_scoped: false,
  zone_scoped: false,
};

const ROSTER = [
  {
    id: "u1",
    idp_user_id: "idp-1",
    email: "ada@localhost",
    display_name: "Ada Lovelace",
    avatar_url: "",
    role: "admin",
    status: "active",
  },
  {
    id: "u2",
    idp_user_id: "idp-2",
    email: "newbie@localhost",
    display_name: "New Bie",
    avatar_url: "",
    role: "staff",
    status: "pending",
  },
];

test.describe("Users — resend invite", () => {
  test("admin resends a pending invite", async ({ page }) => {
    let resentFor: string | null = null;
    await page.route("**/api/v1/me", async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: CAPS });
      else await route.continue();
    });
    await page.route("**/api/v1/users", async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: ROSTER });
      else await route.continue();
    });
    await page.route("**/api/v1/users/stats", async (route) => {
      await route.fulfill({
        json: { members: 2, roles: 5, active_now: 1, pending_invites: 1, twofa_coverage: 50 },
      });
    });
    await page.route("**/api/v1/me/2fa", async (route) => {
      await route.fulfill({ status: 204, body: "" });
    });
    await page.route("**/api/v1/audit*", async (route) => {
      await route.fulfill({ json: [] });
    });
    await page.route("**/api/v1/users/idp-2/resend", async (route) => {
      resentFor = "idp-2";
      await route.fulfill({ status: 204, body: "" });
    });

    await seedOidcSession(page, { role: "admin" });
    await page.goto("/users");
    await page.waitForLoadState("networkidle");

    // Select the pending member, then resend their invite.
    await page.getByText("New Bie").click();
    const resend = page.getByRole("button", { name: "Resend invite" });
    await expect(resend).toBeVisible();
    await resend.click();

    await expect(page.getByRole("button", { name: "Invite resent ✓" })).toBeVisible();
    expect(resentFor).toBe("idp-2");
  });
});
