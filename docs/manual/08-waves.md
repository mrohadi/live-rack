# 8. Waves

**Where:** sidebar → Workflows → **Waves** (`/waves`)

A **wave** lets one person pick **several orders in a single trip**. Instead of
walking the floor once per order, you walk it once for the whole batch.

## 8.1 When to use a wave

Use waves when many small orders need the same or nearby items. Batching cuts
walking time dramatically — the biggest labour saving in the warehouse.

## 8.2 Create a wave

1. First, create the individual [pick lists](07-picking.md) (one per order) and
   leave them **open** (not started).
2. In the Waves screen, click **+ New**.
3. Give the wave a **reference** and **tick at least two open pick lists** to
   batch.
4. Create. live-rack **merges** the lines: for each SKU+zone it sums the
   quantity needed across all the chosen orders, then builds **one shortest
   route** through those merged stops.

## 8.3 Pick the wave

Open the wave to see the **merged route**: one numbered stop per SKU+zone, each
showing the total quantity to grab and how many orders it serves.

1. Click **Start picking**.
2. At each stop, confirm the **combined** quantity (you grab them all at once).
3. live-rack **allocates** what you picked back to the individual orders
   automatically — filling orders in turn (first come, first served).
4. **Complete** the wave.

## 8.4 Short picks in a wave

If you cannot grab the full combined amount, confirm what you got. live-rack
distributes it across the orders in order; orders that miss out are marked
short and a **Restock** task is raised (see [Tasks](06-tasks.md)). On-hand
[Inventory](05-inventory.md) is reduced once for the whole stop.

## 8.5 Hands-free

Like single picking, you can scan SKUs to confirm stops (see
[Scanner](04-scanner.md) and the scan-to-pick section in
[Picking](07-picking.md)).

## 8.6 After a wave

Each member order's pick list is now fulfilled. Pack and ship each on the
[Dispatch](09-dispatch.md) screen.

---

**Connects to:** [Picking](07-picking.md) (member orders) ·
[Map & Zones](03-map-and-zones.md) (merged route) ·
[Inventory](05-inventory.md) (stock drawn down) · [Tasks](06-tasks.md)
(short-pick restock) · [Dispatch](09-dispatch.md) (next step)
