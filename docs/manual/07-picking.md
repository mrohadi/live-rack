# 7. Picking

**Where:** sidebar → Workflows → **Picking** (`/picking`)

Picking is how you collect the items for an order and walk them out of stock.
live-rack builds the **shortest route** through your zones and lets you confirm
each item — by hand or by scan.

## 7.1 Create a pick list

1. Click **+ New** in the Picking screen.
2. Give it a **reference** (e.g. an order number) and add the **SKUs and
   quantities** to collect.
3. Create. For each SKU, live-rack finds the **best zone** to pick from (the
   one holding the most stock), then orders all the stops into the
   **shortest walking path** using your [zone map](03-map-and-zones.md)
   coordinates.

## 7.2 The run view

Open a pick list to see its **route**: a numbered list of stops and a small map
plotting the path. Each stop shows the SKU, the zone, and the quantity needed.

1. Click **Start picking** to begin.
2. At each stop, confirm what you collected. The quantity defaults to the
   amount requested.
3. A **progress bar** tracks how many stops are done.
4. Click **Complete** when finished.

## 7.3 Pick by scan (hands-free)

Instead of typing, turn on **Scan to pick**:

- The camera or your connected USB scanner reads SKUs.
- Each scan credits the matching stop and counts up toward its required
  quantity.
- A scan that does not match any pending stop shows a **mis-scan** warning, so
  you never pick the wrong thing.

## 7.4 Short picks

If a zone does not have enough stock, confirm the amount you actually picked.
live-rack marks the line **short** and automatically raises a **Restock** task
(see [Tasks](06-tasks.md)) for the missing units. On-hand
[Inventory](05-inventory.md) is reduced by exactly what you picked.

## 7.5 What happens on completion

A completed pick list is ready to be **packed and shipped**. Create a shipment
from it on the [Dispatch](09-dispatch.md) screen.

> Picking many orders at once? Batch them into a single trip with
> [Waves](08-waves.md).

---

**Connects to:** [Map & Zones](03-map-and-zones.md) (route) ·
[Inventory](05-inventory.md) (stock drawn down) · [Scanner](04-scanner.md)
(scan-to-pick) · [Tasks](06-tasks.md) (short-pick restock) ·
[Waves](08-waves.md) (batching) · [Dispatch](09-dispatch.md) (next step)
