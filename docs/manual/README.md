# live-rack — Warehouse Operator Manual

A complete, plain-English guide to running your warehouse with live-rack. It
covers every screen, what each does, how to use it, and — most importantly —
**how the screens connect** so goods flow smoothly from arrival to dispatch.

If you are new, read [Getting Started](01-getting-started.md) first, then
[Navigation](02-navigation.md). After that, jump to whichever feature you need.
To see the big picture of how everything fits together, read
[End-to-End Workflows](14-end-to-end-workflows.md).

---

## What live-rack is

live-rack is a multi-tenant SaaS platform for **warehouse zoning, real-time
tracking, fulfilment, and analytics**. One company (an *organization*) can run
multiple **stores** (warehouses). Everything you see is scoped to the store you
have selected and the permissions of your role.

## The warehouse flow at a glance

```
        ┌──────────┐   scan in    ┌───────────┐
        │  Scanner │ ───────────► │ Inventory │
        └──────────┘              └─────┬─────┘
              ▲                          │ low stock
              │ placed on                ▼
        ┌──────────┐               ┌───────────┐
        │   Map    │◄──────────────│   Tasks   │
        │ & Zones  │   zone rules  └───────────┘
        └──────────┘
              │ pick path
              ▼
   ┌──────────┐   batch   ┌──────────┐   pack    ┌───────────┐
   │ Picking  │ ────────► │  Waves   │ ────────► │ Dispatch  │
   └──────────┘           └──────────┘           └───────────┘
              │                                        │
              ▼                                        ▼
        ┌──────────┐                            ┌────────────┐
        │ Pipelines│                            │Integrations│
        │(workflows)│                           │ (carriers) │
        └──────────┘                            └────────────┘

  Everything feeds ► Analytics ◄ and is governed by ► Users & Access
```

## Table of contents

| # | Chapter | What it covers |
|---|---------|----------------|
| 0 | [Introduction](00-introduction.md) | Concepts: org, store, zone, SKU, role |
| 1 | [Getting Started](01-getting-started.md) | Sign up, log in, 2FA, pick a store |
| 2 | [Navigation](02-navigation.md) | Sidebar, top bar, search, notifications, density |
| 3 | [Map & Zones](03-map-and-zones.md) | Draw zones, set rules, capacity |
| 4 | [Scanner](04-scanner.md) | Camera + USB scanning, validation, offline |
| 5 | [Inventory](05-inventory.md) | Stock, reorder, transfer, cycle counts, CSV |
| 6 | [Tasks](06-tasks.md) | Work board, assignments, notifications |
| 7 | [Picking](07-picking.md) | Pick lists, map route, scan-to-pick |
| 8 | [Waves](08-waves.md) | Batch many orders into one route |
| 9 | [Dispatch](09-dispatch.md) | Pack, carrier, tracking, ship out |
| 10 | [Pipelines](10-pipelines.md) | Custom workflows, SLAs, bottlenecks |
| 11 | [Analytics](11-analytics.md) | Heatmaps, velocity, sell-through |
| 12 | [Integrations](12-integrations.md) | Shopify, Square, Stripe, Shippo, more |
| 13 | [Users & Access](13-users-and-access.md) | Roles, 2FA, service tokens, audit |
| 14 | [End-to-End Workflows](14-end-to-end-workflows.md) | How features link in real scenarios |
| 15 | [Glossary](15-glossary.md) | Every term in one place |

---

> **Tip:** Each chapter ends with a **"Connects to"** section listing the other
> screens it links to, so you always know where a piece of work goes next.
