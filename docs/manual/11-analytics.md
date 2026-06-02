# 11. Overview & Analytics

live-rack has two reporting screens: the **Overview** dashboard (your daily
snapshot) and **Analytics** (deeper trends).

## 11.1 Overview (the dashboard)

**Where:** sidebar → Operate → **Overview** (`/`)

This is your landing page. It shows the headline numbers for the selected
store at a glance — for example recent revenue (with a 7-day sparkline), stock
health, and open work — so you know the state of the warehouse the moment you
log in. Use it as a starting point and click through to the detailed screens.

## 11.2 Analytics

**Where:** sidebar → Insights & Setup → **Analytics** (`/analytics`)

Analytics turns the raw activity (scans, sales, movements) into insight.

### Zone heatmap

A colour map of your floor showing **where activity concentrates** — busy pick
faces glow hot, quiet corners stay cool. Overlay it mentally on the
[Map](03-map-and-zones.md): hot zones near the dispatch door are good; hot
zones far away suggest you should re-slot fast movers closer.

### Zone performance

Bars and sparklines comparing zones over time, so you can see which areas are
pulling their weight and which are underused.

### Product metrics

- **Velocity** — how fast each SKU moves (picks over 7 / 30 days), also shown
  on the [Inventory](05-inventory.md) screen.
- **Time-to-sell & sell-through** — how quickly stock turns into sales.
- **Co-purchase lift** — which products tend to sell together (useful for
  slotting them near each other).
- **Demographics** — who is buying, where the data is available.

## 11.3 Where the numbers come from

Analytics is fed by everyday activity: every valid [scan](04-scanner.md),
[inventory](05-inventory.md) change, and sale (via
[Integrations](12-integrations.md)) becomes a data point. You do not enter
anything here — it builds itself from how you run the warehouse.

---

**Connects to:** [Map & Zones](03-map-and-zones.md) (heatmap overlay) ·
[Inventory](05-inventory.md) (velocity & value) ·
[Integrations](12-integrations.md) (sales data) · [Scanner](04-scanner.md)
(activity feed)
