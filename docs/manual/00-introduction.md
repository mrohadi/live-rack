# 0. Introduction & Core Concepts

Before using any screen, it helps to understand five words that appear
everywhere in live-rack.

## Organization

Your company. It is the top-level tenant. All your data — stores, users, stock,
orders — belongs to one organization and is never visible to another. This
isolation is enforced at the database level, so there is no way for one
company's data to leak into another's.

## Store

A single warehouse or location. An organization can have many stores. **You
always work inside one store at a time**, chosen with the store selector in the
top bar (for example, "Store #14"). Switching stores changes every screen to
show that store's zones, stock, tasks, and orders.

## Zone

A physical area on the warehouse floor — a rack, a bay, a cold room, a staging
area. Zones are drawn on the [Map](03-map-and-zones.md) and have:

- **Coordinates and size** — where they sit on the map (used to plan pick routes).
- **Capacity** — the maximum units the zone can hold.
- **Type** — general, frozen, returns, staging, display, or checkout.
- **Rules (constraints)** — which product categories are allowed or denied,
  per-SKU limits, dwell-time rules, and whether a second scan is required.

## SKU & Item

A **SKU** (Stock Keeping Unit) is the unique code for a product, like
`SKU-54345`. An **Item** is the product record behind a SKU: its name,
category, price, reorder point, and status. The same SKU can sit in several
zones; the quantity in each zone is an **item location**.

## Role

What you are allowed to do. Every user has one role per organization:

| Role | Can do |
|------|--------|
| **Admin** | Everything, including users, integrations, billing |
| **Manager** | Run operations, edit zones, manage pipelines & tasks, export reports |
| **Staff** | Scan, move inventory, work their own tasks |
| **Read-only** | View dashboards and export reports; no changes |
| **Service** | A non-human token for integrations (scan + move inventory) |

If a button is missing or you get a "forbidden" message, your role does not
have that permission. See [Users & Access](13-users-and-access.md) for the full
matrix and how to change roles.

---

**Connects to:** [Getting Started](01-getting-started.md) ·
[Map & Zones](03-map-and-zones.md) · [Users & Access](13-users-and-access.md)
