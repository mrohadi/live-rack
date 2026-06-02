# 10. Pipelines

**Where:** sidebar → Workflows → **Pipelines** (`/pipelines`)

Pipelines are **custom, multi-step workflows** for processes that are not a
simple pick-and-ship — for example refurbishing returned goods, quality
inspection, or kitting. Think of a pipeline as a Kanban board tailored to a
warehouse process, with time targets built in.

> Creating or changing pipelines, and moving cards, requires **Manager** or
> **Admin** permission. If buttons are missing or you see "forbidden," ask an
> Admin to raise your role (see [Users & Access](13-users-and-access.md)).

## 10.1 Stages and cards

- A **pipeline** is a series of named **stages** (columns), each with an
  optional **SLA** (a target time a card should spend there).
- A **card** is one unit of work (e.g. a returned item) moving left to right
  through the stages.

## 10.2 Start from a template

The quickest way to begin is a **built-in template**. For example, the
**Item Restoration** template creates a ready-made pipeline with the stages a
refurbishment process needs. Click the template button to instantiate it, then
adjust.

## 10.3 Move work through the board

Drag a card from one stage to the next as the work progresses. The card's time
in a stage is tracked automatically.

## 10.4 Ageing & bottlenecks

This is what makes pipelines powerful:

- A card that sits in a stage **longer than its SLA** is flagged as **ageing**
  (shown in red).
- The board highlights the **bottleneck** — the stage with the most overdue
  cards — so a manager can see exactly where work is piling up.
- Managers are alerted when a stage breaches its SLA, so problems surface
  early.

Use this to spot a stalled refurbishment line or a backed-up inspection step
before it delays shipments.

---

**Connects to:** [Tasks](06-tasks.md) (day-to-day jobs vs. structured
workflows) · [Analytics](11-analytics.md) (throughput) ·
[Users & Access](13-users-and-access.md) (who may edit)
