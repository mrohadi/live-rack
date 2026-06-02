# 15. Glossary

Every term used in this manual, in one place.

| Term | Meaning |
|------|---------|
| **Ageing** | A pipeline card that has stayed in a stage longer than its SLA. Flagged red. |
| **Allocation** | In a wave, splitting a combined pick back across the member orders (first come, first served). |
| **Audit log** | Append-only record of sensitive actions (corrections, picks, dispatches) — who did what, when. |
| **Blind count** | A cycle count where the system quantity is hidden so the physical count stays honest. |
| **Bottleneck** | The pipeline stage with the most overdue (ageing) cards. |
| **Capacity** | The maximum number of units a zone can hold. 0 = unlimited. |
| **Card** | A unit of work moving through a pipeline's stages. |
| **Connector** | A link to an outside system (Shopify, Square, Stripe, Shippo, etc.). |
| **Constraints (zone rules)** | Allowed/denied categories, per-SKU limits, dwell time, dual-scan requirement on a zone. |
| **Cycle count** | A spot-check of physical stock in a zone, reconciled against the system. |
| **Dispatch** | The packing-and-shipping stage; also the screen for it. |
| **Dual scan** | A required second confirming scan before placing certain stock. |
| **Dwell time** | The minimum time before an item may move again. |
| **Item** | The product record behind a SKU (name, category, price, reorder point). |
| **Item location** | The quantity of one SKU in one zone. |
| **Mis-scan** | A scan that breaks a rule or does not match the expected item; rejected with a reason. |
| **Organization (org)** | Your company — the top-level tenant. Data is isolated per org. |
| **Pick list** | The set of SKUs/quantities to collect for an order, with a planned route. |
| **Pipeline** | A custom, multi-stage workflow (e.g. item restoration). |
| **Reorder point** | The stock level at or below which a Restock task is raised automatically. |
| **Role** | What a user may do: Admin, Manager, Staff, Read-only, or Service. |
| **Service token** | A non-human credential for integrations/devices, using the Service role. |
| **Sell-through** | How quickly stock converts into sales. |
| **Shipment** | A packed order with carrier + tracking, moving packing → packed → dispatched. |
| **Short pick** | When a picker cannot fill the requested quantity; raises a Restock task. |
| **SKU** | Stock Keeping Unit — the unique code for a product. |
| **SLA** | A target time a pipeline card should spend in a stage. |
| **Store** | A single warehouse/location. You work in one store at a time. |
| **Two-factor authentication (2FA)** | A second login factor required for high-impact permissions. |
| **Velocity** | How fast a SKU moves (picks over 7 / 30 days). |
| **Wave** | A batch of pick lists picked together as one merged route. |
| **Zone** | A physical area on the warehouse map, with type, capacity, and rules. |

---

**Back to:** [Manual home](README.md)
