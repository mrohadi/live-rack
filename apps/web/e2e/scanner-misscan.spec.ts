import type { Route } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

// The scan endpoint the ScannerPage POSTs to (any origin/base-url).
const SCAN_API = /\/scan(\?|$)/;

const CORS = {
  "access-control-allow-origin": "*",
  "access-control-allow-methods": "POST,OPTIONS",
  "access-control-allow-headers": "authorization,content-type",
};

/** A dwell-violation verdict, mirroring the API ValidateResponse for a mis-scan. */
const BLOCKED = {
  valid: false,
  code: "dwell_violation",
  reason: "zone: item rescanned inside dwell window",
};

const ACCEPTED = { valid: true };

/** Answers the cross-origin preflight; POST gets the supplied verdict. */
async function stubScan(route: Route, verdict: () => unknown) {
  const method = route.request().method();
  if (method === "OPTIONS") {
    await route.fulfill({ status: 204, headers: CORS });
  } else if (method === "POST") {
    await route.fulfill({ json: verdict(), headers: CORS });
  } else {
    await route.continue();
  }
}

test.describe("Scanner — mis-scan blocking with reason", () => {
  test.beforeEach(async ({ page }) => {
    await seedOidcSession(page);
  });

  test("blocks a mis-scan and shows the rejection reason", async ({ page }) => {
    await page.route(SCAN_API, (route) => stubScan(route, () => BLOCKED));

    await page.goto("/scanner");
    await page.waitForLoadState("networkidle");

    // Drive a scan via the manual entry affordance.
    await page.getByTestId("manual-sku-input").fill("SKU-DWELL-1");
    await page.getByTestId("manual-scan-btn").click();

    // The block banner appears and carries the server-supplied reason.
    const banner = page.getByTestId("scan-blocked");
    await expect(banner).toBeVisible();
    await expect(banner).toContainText("SKU-DWELL-1");
    await expect(page.getByTestId("scan-blocked-reason")).toHaveText(BLOCKED.reason);

    // A blocked scan is NOT added to the accepted list.
    await expect(page.getByTestId("accepted-scans")).not.toContainText("SKU-DWELL-1");
  });

  test("accepts a valid scan and clears any prior block", async ({ page }) => {
    let verdict: typeof BLOCKED | typeof ACCEPTED = BLOCKED;
    await page.route(SCAN_API, (route) => stubScan(route, () => verdict));

    await page.goto("/scanner");
    await page.waitForLoadState("networkidle");

    // First scan is blocked.
    await page.getByTestId("manual-sku-input").fill("SKU-BAD");
    await page.getByTestId("manual-scan-btn").click();
    await expect(page.getByTestId("scan-blocked")).toBeVisible();

    // Next scan is valid — block clears, SKU lands in the accepted list.
    verdict = ACCEPTED;
    await page.getByTestId("manual-sku-input").fill("SKU-GOOD");
    await page.getByTestId("manual-scan-btn").click();

    await expect(page.getByTestId("scan-blocked")).toHaveCount(0);
    await expect(page.getByTestId("accepted-scans")).toContainText("SKU-GOOD");
  });
});
