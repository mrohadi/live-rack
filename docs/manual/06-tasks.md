# 6. Tasks

**Where:** sidebar → Workflows → **Tasks** (`/tasks`)

Tasks are the warehouse's shared to-do board. Anything that needs doing —
restock this SKU, check that zone, prep an order — lives here as a card.

## 6.1 The board

Tasks are shown as cards in columns by status (for example **To do →
In progress → Done**). Drag a card between columns as work progresses, or open
a card to change it.

A task carries:

- **Title** — what to do (e.g. "Restock SKU-54345").
- **Status** — where it is in the flow.
- **Priority** — low, medium, or high.
- **Zone** — the area it relates to (optional).
- **Assignee** — the person responsible (optional).

## 6.2 Create a task

Click **+ New task**, give it a title and priority, optionally a zone and an
assignee, then save. It appears on the board for everyone in the store.

## 6.3 Assignments & notifications

When you **assign** a task to someone (at creation or later), that person gets
a [notification](02-navigation.md) under their bell. They are also notified when
an assigned task is **due soon**. Notifications are personal — each user only
sees their own.

## 6.4 Where tasks come from automatically

You do not have to create every task by hand. live-rack raises tasks for you:

- **Auto-restock:** when a pick or scan drops stock to/below its reorder point,
  a high-priority **Restock** task appears (see [Inventory](05-inventory.md)).
- **Short picks:** if a picker cannot fill an order line, a restock task is
  raised for the missing SKU (see [Picking](07-picking.md) and
  [Waves](08-waves.md)).
- **Insights:** recommendations from [Integrations & Analytics](12-integrations.md)
  can be **applied** to create a task (for example, "move umbrellas to the
  front — rain forecast").

To avoid duplicates, live-rack will not open a second open restock task for a
SKU that already has one.

---

**Connects to:** [Inventory](05-inventory.md) & [Scanner](04-scanner.md)
(auto-restock) · [Picking](07-picking.md)/[Waves](08-waves.md) (short-pick
tasks) · [Integrations](12-integrations.md) (apply a recommendation) ·
[Navigation](02-navigation.md) (notifications)
