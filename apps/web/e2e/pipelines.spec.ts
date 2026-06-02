import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const LIST = "**/api/v1/stores/*/pipelines";
const BOARD = "**/api/v1/stores/*/pipelines/p1";
const CARD = "**/api/v1/stores/*/pipelines/p1/cards/*";

const PIPELINES = [{ id: "p1", key: "item-restoration", name: "Item Restoration" }];

const BOARD_DATA = {
  pipeline: PIPELINES[0],
  stages: [
    { position: 0, name: "Intake", sla_seconds: 3600 },
    { position: 1, name: "Repair", sla_seconds: 3600 },
  ],
  cards: [
    {
      id: "card-1111-2222-3333-4444",
      stage_position: 0,
      title: "Refinish dresser",
      priority: "high",
      entered_stage_at: new Date().toISOString(),
      age_seconds: 120,
      ageing: false,
    },
  ],
};

test.describe("Pipelines — board + card move", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(LIST, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: PIPELINES });
      else await route.continue();
    });
    await page.route(BOARD, async (route) => {
      if (route.request().method() === "GET") await route.fulfill({ json: BOARD_DATA });
      else await route.continue();
    });
    await page.route(CARD, async (route) => {
      if (route.request().method() === "PATCH") {
        const body = route.request().postDataJSON();
        await route.fulfill({ json: { ...BOARD_DATA.cards[0], ...body } });
      } else {
        await route.continue();
      }
    });

    await seedOidcSession(page);
    await page.goto("/pipelines");
    await page.waitForLoadState("networkidle");
  });

  test("renders stages and moves a card to the next stage", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Pipelines" })).toBeVisible();
    await expect(page.getByTestId("stage-0")).toBeVisible();
    await expect(page.getByTestId("stage-1")).toBeVisible();

    const card = page.getByTestId("pipeline-card");
    await expect(card).toBeVisible();

    const cardBox = await card.boundingBox();
    const targetBox = await page.getByTestId("stage-1").boundingBox();
    if (!cardBox || !targetBox) throw new Error("missing layout box");

    const patchPromise = page.waitForRequest(
      (r) => r.url().includes("/cards/") && r.method() === "PATCH",
    );

    // dnd-kit PointerSensor needs >6px movement.
    await page.mouse.move(cardBox.x + cardBox.width / 2, cardBox.y + cardBox.height / 2);
    await page.mouse.down();
    await page.mouse.move(targetBox.x + targetBox.width / 2, targetBox.y + 40, { steps: 12 });
    await page.mouse.up();

    const patch = await patchPromise;
    expect(patch.postDataJSON()).toEqual({ stage_position: 1 });
  });
});
