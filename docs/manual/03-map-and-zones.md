# 3. Map & Zones

**Where:** sidebar → Operate → **Map & Zones** (`/map`)

The Map is your warehouse floor plan. Everything physical in live-rack — where
stock sits, how pickers walk, what may go where — starts here.

## 3.1 What a zone is

A **zone** is a rectangle on the floor plan representing a real area: a rack
bay, a cold room, a returns corner, a staging lane. Each zone has a position,
a size, a colour, and a set of rules.

## 3.2 Drawing and arranging zones

1. On the Map, **draw** a new zone by dragging a rectangle, or add one and
   resize it.
2. **Move** a zone by dragging it; **resize** with its handles.
3. Give it a clear **name** (e.g. "A1 — Ambient", "FRZ — Freezer").

The zone's position matters: live-rack uses the coordinates to **plan the
shortest pick route** later in [Picking](07-picking.md) and
[Waves](08-waves.md).

## 3.3 Zone types

Choose a type that matches the area's purpose:

| Type | Typical use |
|------|-------------|
| **General** | Everyday ambient storage |
| **Frozen** | Cold/freezer stock |
| **Returns** | Items coming back from customers |
| **Staging** | Goods waiting to be picked or shipped |
| **Display** | Show-floor or pick-face |
| **Checkout** | Dispatch/exit point |

## 3.4 Capacity

Set a **capacity** (maximum units) for each zone. When someone tries to place
more than the zone can hold — by scanning or transferring — live-rack blocks it
with a "capacity exceeded" message. A capacity of 0 means unlimited.

## 3.5 Zone rules (constraints)

Rules keep the wrong stock out of the wrong place. For each zone you can set:

- **Allowed categories** — only these product categories may enter.
- **Denied categories** — these may never enter (denied always wins over
  allowed).
- **Max units per SKU** — cap how much of a single product sits here.
- **Require dual scan** — a second confirming scan before placement (for
  high-value or hazardous goods).
- **Dwell time** — minimum time before an item may move again.

These rules are enforced live by the [Scanner](04-scanner.md): a scan that
breaks a rule is rejected with a reason, and nothing moves.

## 3.6 Reading the map day-to-day

The Map also reflects **real-time activity**: as scans happen, zones update so
you can see where work is occurring. Combined with the
[Analytics heatmap](11-analytics.md), the Map shows both the *layout* and the
*performance* of your floor.

---

**Connects to:** [Scanner](04-scanner.md) (enforces zone rules) ·
[Inventory](05-inventory.md) (stock per zone) ·
[Picking](07-picking.md) & [Waves](08-waves.md) (routes use zone coordinates) ·
[Analytics](11-analytics.md) (zone heatmap)
