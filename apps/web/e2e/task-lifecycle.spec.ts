import { expect, test } from "@playwright/test";
import { seedOidcSession } from "./auth";

const STORE_ID = "22222222-2222-2222-2222-222222222222";
const API = `/api/v1/stores/${STORE_ID}/tasks`;

const TASKS = [
  {
    id: "11111111-0000-0000-0000-000000000001",
    store_id: STORE_ID,
    title: "Restock aisle A1",
    status: "todo",
    priority: "high",
    updated_at: new Date().toISOString(),
  },
  {
    id: "11111111-0000-0000-0000-000000000002",
    store_id: STORE_ID,
    title: "Audit returns bin",
    status: "in_progress",
    priority: "med",
    updated_at: new Date().toISOString(),
  },
];

test.describe("Tasks — kanban lifecycle", () => {
  test.beforeEach(async ({ page }) => {
    let tasks = [...TASKS];
    await page.route(`**${API}`, async (route) => {
      if (route.request().method() === "GET") {
        await route.fulfill({ json: tasks });
      } else {
        await route.continue();
      }
    });
    // PATCH /tasks/:id → echo back the moved task
    await page.route(`**${API}/*`, async (route) => {
      if (route.request().method() === "PATCH") {
        const body = route.request().postDataJSON();
        const id = route.request().url().split("/").pop()!;
        tasks = tasks.map((t) => (t.id === id ? { ...t, ...body } : t));
        await route.fulfill({ json: tasks.find((t) => t.id === id) });
      } else {
        await route.continue();
      }
    });

    await seedOidcSession(page);
    await page.goto("/tasks");
    await page.waitForLoadState("networkidle");
  });

  test("renders columns and moves a card from To do to Done", async ({ page }) => {
    await expect(page.getByTestId("column-todo")).toBeVisible();
    await expect(page.getByTestId("column-done")).toBeVisible();

    const card = page.locator('[data-task-id="11111111-0000-0000-0000-000000000001"]');
    await expect(card).toBeVisible();
    // Card starts in To do
    await expect(page.getByTestId("column-todo").getByTestId("task-card")).toHaveCount(1);

    const cardBox = await card.boundingBox();
    const doneBox = await page.getByTestId("column-done").boundingBox();
    if (!cardBox || !doneBox) throw new Error("missing layout box");

    const patchPromise = page.waitForRequest(
      (r) => r.url().includes("/tasks/") && r.method() === "PATCH",
    );

    // dnd-kit PointerSensor needs >6px movement; drag in steps to the Done column.
    await page.mouse.move(cardBox.x + cardBox.width / 2, cardBox.y + cardBox.height / 2);
    await page.mouse.down();
    await page.mouse.move(doneBox.x + doneBox.width / 2, doneBox.y + 40, { steps: 12 });
    await page.mouse.up();

    const patch = await patchPromise;
    expect(patch.postDataJSON()).toEqual({ status: "done" });

    // Optimistic move lands the card in Done.
    await expect(page.getByTestId("column-done").getByTestId("task-card")).toHaveCount(1);
  });
});
