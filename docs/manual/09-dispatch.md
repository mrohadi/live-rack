# 9. Dispatch (Packing & Shipping)

**Where:** sidebar → Workflows → **Dispatch** (`/shipments`)

Dispatch is the last step: turn picked goods into a **shipment**, record the
carrier and tracking number, and send it out.

## 9.1 The shipment lifecycle

A shipment moves through four states:

```
packing ──► packed ──► dispatched
   │            │
   └── cancel ──┘   (cancel allowed until it ships)
```

## 9.2 Create a shipment from a completed pick list

1. Click **+ New** in the Dispatch screen.
2. Give it a **reference** and select a **completed** [pick list](07-picking.md).
   (Only completed pick lists appear — you cannot ship goods you have not yet
   picked.)
3. Create. live-rack **snapshots the picked lines** into the shipment as its
   packing list (every SKU with a picked quantity).

## 9.3 Pack

In the shipment detail, review the item list, physically pack the parcel, then
click **Mark packed**. The shipment moves to **packed** and is ready to leave.

## 9.4 Dispatch

Once packed, enter the **carrier** (e.g. UPS) and the **tracking number**, then
click **Dispatch**. The shipment is timestamped and marked **dispatched**, and
the carrier + tracking are saved for reference. This action is recorded in the
[audit trail](13-users-and-access.md).

## 9.5 Cancel

Before a shipment ships, you can **Cancel** it (while packing or packed). Once
dispatched, it can no longer be cancelled.

## 9.6 Tracking updates

Carrier tracking events (via [Integrations](12-integrations.md), e.g. Shippo)
flow back into live-rack so you can follow a parcel after it leaves.

---

**Connects to:** [Picking](07-picking.md)/[Waves](08-waves.md) (source of
goods) · [Integrations](12-integrations.md) (carrier tracking) ·
[Users & Access](13-users-and-access.md) (dispatch is audited)
