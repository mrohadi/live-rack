# 5. Inventory

**Where:** sidebar → Operate → **Inventory** (`/inventory`)

The Inventory screen is the single source of truth for **what you have and
where it is**. Each row is a SKU in a zone, with its quantity, movement, and
value.

## 5.1 Reading the table

| Column | Meaning |
|--------|---------|
| **SKU / Name** | The product |
| **Zone** | Where this stock sits |
| **Stock** | On-hand quantity, with a status: In stock / Low / Out |
| **Velocity** | How fast it moves (picks over the last 7 / 30 days) |
| **Value** | Quantity × unit price |

Use the **stock tabs** to filter to All, In stock, Low, or Out. The **stock
status** is calculated from the quantity versus the item's **reorder point**.

## 5.2 Add an item

Click **+ Add item**, then enter the SKU, name, category, price, reorder point,
and starting quantity/zone. The item appears in the table immediately.

## 5.3 Edit an item / correct a quantity

- Click a row to open its **detail drawer**. There you see the item across all
  its zones, recent scan history, and totals.
- **Edit** the item's details (name, category, price, reorder point).
- **Correct the quantity** when a number is wrong (damage, shrinkage). This is
  an absolute correction and is recorded in the [audit trail](13-users-and-access.md).

## 5.4 Transfer stock between zones

Use **Transfer** (the ⇄ button on a row) to move units from one zone to
another. live-rack checks the destination zone's
[rules and capacity](03-map-and-zones.md) first; if the source does not have
enough stock, the transfer is refused.

## 5.5 Reorder points & auto-restock

Set a **reorder point** on each item. When stock falls to or below it — usually
from a pick — live-rack raises a **Restock** task automatically (see
[Tasks](06-tasks.md)). This keeps shelves filled without manual watching.

## 5.6 Cycle counts (blind counting)

Cycle counting verifies physical stock against the system without bias:

1. Click **Cycle count** and pick a zone. live-rack snapshots the SKUs there.
2. The counter enters the **physical count for each SKU**. The system quantity
   is **hidden** during counting (this is what "blind" means), so counts stay
   honest.
3. **Complete** the count. live-rack shows the **variance** per SKU
   (shrinkage in red, surplus in blue) and reconciles the on-hand numbers to
   what you counted. Every correction is written to the audit trail.

## 5.7 Export to CSV

Click **Export CSV** to download the current inventory view as a spreadsheet
(properly quoted, opens cleanly in Excel/Sheets) for offline reporting or
sharing.

---

**Connects to:** [Scanner](04-scanner.md) (updates stock) ·
[Map & Zones](03-map-and-zones.md) (zones & rules) ·
[Tasks](06-tasks.md) (auto-restock) · [Picking](07-picking.md) (picks draw
from stock) · [Analytics](11-analytics.md) (velocity & value)
