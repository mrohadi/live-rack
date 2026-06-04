import { setupClerkTestingToken } from "@clerk/testing/playwright";
import { expect, test } from "@playwright/test";

const STORE_ID = "22222222-2222-2222-2222-222222222222";
const API = `/api/v1/stores/${STORE_ID}/zones`;

/** Two seeded zones the DOM map renders as `[data-testid=zone-box]`. */
const ZONE_A = {
  id: "aaaaaaaa-0000-0000-0000-000000000001",
  org_id: "11111111-1111-1111-1111-111111111111",
  store_id: STORE_ID,
  name: "Apparel A2",
  type: "general",
  x: 20,
  y: 20,
  width: 24,
  height: 20,
  color: "#2563eb",
  items: 40,
  capacity: 100,
  constraints: {},
};

const ZONE_B = {
  ...ZONE_A,
  id: "aaaaaaaa-0000-0000-0000-000000000002",
  name: "Cold Chain B1",
  type: "frozen",
  x: 60,
  y: 20,
  items: 95,
  capacity: 100,
};

test.describe("Map — DOM zone editor", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(`**/api/v1/stores/${STORE_ID}/inventory*`, (route) =>
      route.fulfill({ json: [] }),
    );
    await seedOidcSession(page);
  });

  test("switches views, selects, drags, filters, and deletes a zone", async ({ page }) => {
    let zones = [ZONE_A, ZONE_B];

    await page.route(`**${API}/**`, async (route) => {
      const method = route.request().method();
      if (method === "PUT") {
        const body = route.request().postDataJSON();
        zones = zones.map((z) => (z.id === body.id ? { ...z, ...body } : z));
        await route.fulfill({ json: body });
      } else if (method === "DELETE") {
        const id = route.request().url().split("/").pop();
        zones = zones.filter((z) => z.id !== id);
        await route.fulfill({ status: 204, body: "" });
      } else {
        await route.continue();
      }
    });

    // Create sign-in token via Clerk Backend API (Node.js side)
    const tokenRes = await fetch("https://api.clerk.com/v1/sign_in_tokens", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${process.env.CLERK_SECRET_KEY!}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        user_id: process.env.CLERK_E2E_USER_ID!,
        expires_in_seconds: 60,
      }),
    });
    const { token } = (await tokenRes.json()) as { token: string };

    // Navigate to sign-in with token — no UI flow needed
    await setupClerkTestingToken({ page });
    await page.goto(`/sign-in?__clerk_ticket=${token}`);
    await page.waitForURL((url) => !url.pathname.includes("sign-in"), { timeout: 15_000 });

    await page.goto("/map");
    await page.waitForLoadState("networkidle");

    // ── Both zones render ──────────────────────────────────────────────
    const boxes = page.getByTestId("zone-box");
    await expect(boxes).toHaveCount(2);

    // ── View switch ────────────────────────────────────────────────────
    await page.getByRole("button", { name: "Heat", exact: true }).click();
    await page.getByRole("button", { name: "Items", exact: true }).click();
    await page.getByRole("button", { name: "Zones", exact: true }).click();

    // ── Select → sidebar reflects the zone ─────────────────────────────
    await page.getByText("Apparel A2").first().click();
    await expect(page.getByText("zone aaaaaaaa-0000-0000-0000-000000000001")).toBeVisible();

    // ── Drag right → PUT persists a larger x ───────────────────────────
    const target = boxes.first();
    const box = await target.boundingBox();
    if (!box) throw new Error("zone-box not found");
    const fromX = box.x + box.width / 2;
    const fromY = box.y + box.height / 2;

    const putPromise = page.waitForRequest(
      (r) => r.url().includes("/zones/") && r.method() === "PUT",
    );
    await page.mouse.move(fromX, fromY);
    await page.mouse.down();
    await page.mouse.move(fromX + 120, fromY, { steps: 12 });
    await page.mouse.up();

    const putReq = await putPromise;
    const body = putReq.postDataJSON();
    expect(body.id).toBe(ZONE_A.id);
    expect(body.x).toBeGreaterThan(ZONE_A.x);

    // ── Filter by fill (high > 85%) keeps only Cold Chain B1 ───────────
    await page.getByRole("button", { name: /^Filter/ }).click();
    await page.locator("select").last().selectOption("high");
    await expect(page.getByTestId("zone-box")).toHaveCount(1);
    await expect(page.getByText("Cold Chain B1")).toBeVisible();
    // Clear filters back to two zones
    await page.getByRole("button", { name: "Clear filters" }).click();
    await expect(page.getByTestId("zone-box")).toHaveCount(2);

    // ── Delete via ⋯ menu → confirm modal → DELETE fires ───────────────
    await page.getByText("Apparel A2").first().click();
    const delPromise = page.waitForRequest(
      (r) => r.url().includes("/zones/") && r.method() === "DELETE",
    );
    await page.getByRole("button", { name: "more" }).click();
    await page.getByRole("button", { name: "Delete" }).click();
    // Confirm modal appears; click its Delete button.
    const dialog = page.getByRole("dialog", { name: "Delete zone" });
    await expect(dialog).toBeVisible();
    await dialog.getByRole("button", { name: "Delete" }).click();
    await delPromise;
    await expect(page.getByText("Zone deleted")).toBeVisible();
    await expect(page.getByTestId("zone-box")).toHaveCount(1);
  });
});
