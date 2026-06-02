# 4. Scanner

**Where:** sidebar → Operate → **Scanner** (`/scanner`)

The Scanner is how stock physically moves in live-rack. Every place, pick,
move, or count is a **scan**, and every scan is checked against the
[zone's rules](03-map-and-zones.md) before anything changes.

## 4.1 Two ways to scan

- **Camera (phone/tablet/laptop):** the Scanner page opens your camera and
  reads barcodes automatically. Point at the label.
- **USB scanner (Zebra):** connect a handheld Zebra scanner over USB. Click
  **Connect scanner** once; after that, every trigger pull feeds the app
  directly.

Both methods produce the same result: a SKU is captured and validated.

## 4.2 Scan actions

A scan has an **action** that tells live-rack what is happening:

| Action | Meaning |
|--------|---------|
| **Place** | Add stock into a zone |
| **Pick** | Remove stock from a zone (for an order) |
| **Move** | Relocate stock between zones |
| **Count** | Verify quantity during a count |

## 4.3 Validation — why a scan might be rejected

Before stock moves, live-rack checks the target zone's rules. A scan is
**blocked** (with a clear reason) if it would:

- put a product in a zone that **denies** its category,
- put a product in a zone whose **allowed** list excludes it,
- exceed the zone's **capacity**,
- exceed the **max units per SKU** for that zone,
- break a **dwell-time** rule (moved again too soon),
- skip a **required second scan** (dual-scan zones).

A valid scan updates [Inventory](05-inventory.md) immediately and shows on the
[Map](03-map-and-zones.md). A rejected scan changes nothing and tells you why,
so you can place the item correctly.

## 4.4 Working offline

Warehouses have dead spots. If the network drops, the Scanner **queues your
scans on the device** and keeps working. When the connection returns, the queue
**syncs automatically**. You never lose a scan.

## 4.5 Auto-restock

When a pick drops a SKU's on-hand quantity to or below its **reorder point**,
live-rack automatically raises a high-priority **Restock** task on the
[Tasks](06-tasks.md) board — so replenishment happens without anyone
remembering to ask.

---

**Connects to:** [Map & Zones](03-map-and-zones.md) (rules enforced here) ·
[Inventory](05-inventory.md) (updated by every valid scan) ·
[Tasks](06-tasks.md) (auto-restock) · [Analytics](11-analytics.md) (scan
velocity feeds heatmaps)
