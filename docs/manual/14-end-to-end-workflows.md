# 14. End-to-End Workflows

This chapter shows how the screens **link together** in real situations. Follow
any scenario top to bottom to see goods (and data) flow through live-rack.

## 14.1 Set up a new warehouse (one-time)

1. **Sign in** and select your store — [Getting Started](01-getting-started.md).
2. **Draw your zones** on the [Map](03-map-and-zones.md): name them, set
   capacity, set type, add rules (allowed categories, dual-scan, etc.).
3. **Add your products** on [Inventory](05-inventory.md): SKU, name, category,
   price, and **reorder point**.
4. **Add your team** on [Users & Access](13-users-and-access.md): invite users,
   set roles, assign stores; enable 2FA for admins.
5. **Connect your systems** on [Integrations](12-integrations.md): Shopify,
   Square, Stripe, Shippo.

## 14.2 Goods arrive → on the shelf

1. Receiver opens the [Scanner](04-scanner.md) and scans items **into** a zone
   (action: place).
2. The Scanner checks the [zone's rules and capacity](03-map-and-zones.md);
   valid scans update [Inventory](05-inventory.md) instantly and show on the
   [Map](03-map-and-zones.md).
3. A rejected scan explains why (wrong category, full zone), so the item is
   placed correctly.

## 14.3 Order comes in → out the door (single order)

```
Sale (Integrations) ─► Pick list (Picking) ─► Pick the route ─► Shipment (Dispatch) ─► Carrier
```

1. An order arrives from [Integrations](12-integrations.md) (Shopify/Square).
2. Create a [pick list](07-picking.md) for it; live-rack plans the shortest
   route over your zones.
3. Pick it — by hand or **scan-to-pick** ([Scanner](04-scanner.md)). Stock in
   [Inventory](05-inventory.md) drops as you go.
4. **Complete** the pick list.
5. On [Dispatch](09-dispatch.md), create a shipment from the completed pick
   list, **mark packed**, enter carrier + tracking, **dispatch**.
6. Tracking updates flow back via [Integrations](12-integrations.md).

## 14.4 Many orders at once (busy day)

1. Create a [pick list](07-picking.md) per order; leave them open.
2. On [Waves](08-waves.md), batch two or more into a wave. live-rack merges the
   lines and builds **one** route.
3. Walk the route once, confirming the combined quantity per stop (scan or
   type). live-rack **splits** the picks back to each order.
4. Pack and [dispatch](09-dispatch.md) each order.

## 14.5 Stock runs low → automatically refilled

1. A pick drops a SKU to or below its **reorder point**
   ([Inventory](05-inventory.md)).
2. live-rack raises a high-priority **Restock** task
   ([Tasks](06-tasks.md)) and notifies the assignee.
3. A worker restocks the zone (scan: place), clearing the shortage; the task is
   marked done.

## 14.6 Verify stock is accurate (cycle count)

1. On [Inventory](05-inventory.md), start a **Cycle count** for a zone.
2. Count physically and enter numbers **blind** (system quantity hidden).
3. Complete to see **variances** and auto-reconcile; corrections hit the
   [audit trail](13-users-and-access.md).

## 14.7 Refurbish returns (custom workflow)

1. On [Pipelines](10-pipelines.md), start the **Item Restoration** template.
2. Move each returned item's card stage by stage (inspect → repair → restock).
3. Watch for **ageing** cards and the **bottleneck** flag; managers are alerted
   when a stage runs over its SLA.

## 14.8 React to outside signals

1. A weather or transit signal produces a recommendation on
   [Integrations](12-integrations.md).
2. **Apply** it to create a [Task](06-tasks.md) (e.g. "front the umbrellas").
3. The team acts; results show later in [Analytics](11-analytics.md).

## 14.9 Review performance (weekly)

1. Start at the [Overview](11-analytics.md) dashboard for headline numbers.
2. Open [Analytics](11-analytics.md): heatmap (where work happens), zone
   performance, velocity, sell-through, co-purchase.
3. Act on it — re-slot fast movers near dispatch on the
   [Map](03-map-and-zones.md), adjust reorder points in
   [Inventory](05-inventory.md).

---

**Connects to:** every chapter — this is the glue. For definitions, see the
[Glossary](15-glossary.md).
