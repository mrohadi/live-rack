# 13. Users & Access

**Where:** sidebar → Insights & Setup → **Users & Access** (`/users`) — **Admin
only**

This screen manages the people who use live-rack, what they can do, and the
security around it.

## 13.1 Roles and what they can do

Every user has one role per organization. This is the full permission matrix:

| Permission | Admin | Manager | Staff | Read-only | Service |
|------------|:----:|:------:|:----:|:--------:|:------:|
| View dashboards | ✅ | ✅ | ✅ | ✅ | |
| Edit zones | ✅ | ✅ | | | |
| Approve mis-scans | ✅ | ✅ | | | |
| Manage pipelines | ✅ | ✅ | | | |
| Run scanner | ✅ | ✅ | ✅ | | ✅ |
| Move inventory | ✅ | ✅ | ✅ | | ✅ |
| Manage all tasks | ✅ | ✅ | | | |
| Manage own tasks | ✅ | ✅ | ✅ | | |
| Edit users | ✅ | | | | |
| Manage integrations | ✅ | | | | |
| Export reports | ✅ | ✅ | | ✅ | |

> If a button is missing or you get **"forbidden,"** your role lacks that
> permission. Common case: creating a pipeline needs **Manager** or **Admin**.
> An Admin can change a role here.

## 13.2 Invite and manage users

Admins can **invite** new users by email, assign their **role**, and set which
**stores** they may access. Users are provisioned on first login.

## 13.3 Change someone's role

Open the user on this screen and set a new role. Always change roles here rather
than by other means, so the change is tracked and consistent.

## 13.4 Two-factor authentication (2FA)

High-impact permissions — **editing users** and **managing integrations** —
require a verified **second factor**. Set up 2FA with an authenticator app from
this screen. After enabling, you are prompted for a code at login.

## 13.5 Service tokens

For machine-to-machine access (a script, an integration, a device), create a
**service token** instead of sharing a person's login. A service token uses the
restricted **Service** role (scan + move inventory only). Revoke a token any
time.

## 13.6 Audit trail

Sensitive actions are written to an append-only **audit log** — for example
inventory corrections, cycle-count reconciliations, picks, and dispatches. This
gives you an accountable history of who changed what and when.

## 13.7 Multi-tenancy & data isolation

Your organization's data is strictly separated from every other organization at
the database level. Within your org, access is further scoped by **store** and
by **role**, so people only see and touch what they should.

---

**Connects to:** every screen (roles gate what you can do) ·
[Integrations](12-integrations.md) (2FA-gated) · [Inventory](05-inventory.md) &
[Dispatch](09-dispatch.md) (audited actions)
