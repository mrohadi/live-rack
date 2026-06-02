# 12. Integrations

**Where:** sidebar → Insights & Setup → **Integrations** (`/integrations`)

Integrations connect live-rack to the outside systems that drive your
warehouse: sales channels, payments, shipping, and marketing. They keep stock,
orders, and shipping in sync without manual re-keying.

> Managing integrations is a high-impact action and may require **two-factor
> authentication** and **Admin** permission.

## 12.1 Connectors

The **Connectors** panel lists the systems you can link:

| Connector | What it brings in |
|-----------|-------------------|
| **Shopify** | Online orders and sales |
| **Square** | In-person/POS sales |
| **Stripe** | Payments (and platform billing) |
| **Shippo** | Carrier shipping & tracking updates |
| **Klaviyo** | Marketing/customer events |
| **NetSuite** | ERP/accounting (skeleton) |
| **Weather / Transit** | External signals that drive recommendations |

Connect a system once; from then on its events flow into live-rack.

## 12.2 Webhook event log

The **Webhook event log** shows incoming events from connected systems in real
time — useful to confirm a connector is working and to troubleshoot. Each
inbound event is verified for authenticity before it is accepted.

## 12.3 From signal to action — recommendations

Some integrations do more than sync data; they generate **recommendations**.
For example, a weather feed predicting rain, or a transit feed showing a busy
day, produces a **signal card**. You can **Apply** a recommendation, which
creates a [Task](06-tasks.md) (e.g. "move umbrellas to the front display") so
the insight turns into real work on the floor.

## 12.4 How sales become analytics

Orders from Shopify/Square become sales events that power
[Analytics](11-analytics.md) (sell-through, co-purchase, demographics) and can
trigger replenishment via [Tasks](06-tasks.md).

---

**Connects to:** [Analytics](11-analytics.md) (sales data) ·
[Tasks](06-tasks.md) (apply recommendations) · [Dispatch](09-dispatch.md)
(carrier tracking) · [Users & Access](13-users-and-access.md) (2FA-gated)
