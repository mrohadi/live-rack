import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const USERS = "**/api/v1/users";
const ME = "**/api/v1/me";

const ROSTER = [
  {
    id: "u1",
    email: "ada@localhost",
    display_name: "Ada Lovelace",
    avatar_url: "",
    role: "admin",
  },
  {
    id: "u2",
    email: "grace@localhost",
    display_name: "Grace Hopper",
    avatar_url: "",
    role: "staff",
  },
];

const CAPS = {
  user_id: "u1",
  role: "admin",
  mfa_verified: true,
  permissions: ["users.edit"],
  store_scoped: false,
  zone_scoped: false,
};

test.describe("Users — roster + permission matrix", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(ME, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: CAPS });
      else await route.continue();
    });
    await page.route(USERS, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: ROSTER });
      else await route.continue();
    });
    await page.route("**/api/v1/users/stats", async (route) => {
      await route.fulfill({
        json: { members: 2, roles: 5, active_now: 1, pending_invites: 0, twofa_coverage: 50 },
      });
    });
    await page.route("**/api/v1/me/2fa", async (route) => {
      await route.fulfill({ status: 204, body: "" });
    });
    await page.route("**/api/v1/audit*", async (route) => {
      await route.fulfill({ json: [] });
    });

    await seedOidcSession(page, { role: "admin" });
    await page.goto("/users");
    await page.waitForLoadState("networkidle");
  });

  test("renders the roster, caller role, and permission matrix", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Users & Access" })).toBeVisible();

    // Roster rows (Ada also appears in the side detail panel → first()).
    await expect(page.getByText("Ada Lovelace").first()).toBeVisible();
    await expect(page.getByText("Grace Hopper")).toBeVisible();
    await expect(page.getByTestId("user-row")).toHaveCount(2);

    // Caller capabilities surface in the subheader.
    await expect(page.getByText(/you: admin · 2FA on/)).toBeVisible();

    // Static permission matrix rows.
    await expect(page.getByText("View dashboards")).toBeVisible();
    await expect(page.getByText("Manage integrations")).toBeVisible();
  });
});
