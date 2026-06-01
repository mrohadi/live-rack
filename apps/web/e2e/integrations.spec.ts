import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const LIST = "**/api/v1/integrations";
const LOG = "**/api/v1/integrations/webhooks";

const INTEGRATIONS = [
  { id: "int-1", kind: "shopify", status: "connected", external_id: "shop_42" },
  { id: "int-2", kind: "stripe", status: "error", external_id: "acct_99" },
];

const EVENTS = [
  {
    id: "evt-1",
    provider: "shopify",
    event_id: "wh_abc123",
    topic: "orders/create",
    status: "processed",
    received_at: new Date().toISOString(),
  },
];

test.describe("Integrations — connectors + webhook log", () => {
  test.beforeEach(async ({ page }) => {
    // Order matters: register the more specific /webhooks route first.
    await page.route(LOG, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: EVENTS });
      else await route.continue();
    });
    await page.route(LIST, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: INTEGRATIONS });
      else await route.continue();
    });

    await seedOidcSession(page);
    await page.goto("/integrations");
    await page.waitForLoadState("networkidle");
  });

  test("renders webhook config, connectors, and the event log", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Integrations" })).toBeVisible();
    await expect(page.getByTestId("webhook-config")).toBeVisible();

    // Connector cards from the stubbed list (external ids are unique on the page).
    await expect(page.getByText("shop_42")).toBeVisible();
    await expect(page.getByText("acct_99")).toBeVisible();

    // Webhook event log row.
    await expect(page.getByText("orders/create")).toBeVisible();
    await expect(page.getByText("wh_abc123")).toBeVisible();
  });
});
