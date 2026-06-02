# 1. Getting Started

This chapter takes you from no account to a working session inside a store.

## 1.1 Create an account (sign up)

1. Open the app and choose **Sign up** (`/signup`).
2. Enter your work email and a strong password.
3. Submit. live-rack sends a verification email.

## 1.2 Verify your email

1. Open the email and click the verification link, or enter the code on the
   **Verify email** screen (`/verify-email`).
2. Once verified, your account is provisioned into your organization
   automatically on first login.

## 1.3 Log in

1. Go to **Log in** (`/login`).
2. Enter your email and password. live-rack uses secure single sign-on
   (OIDC) behind the scenes.
3. If your account has **two-factor authentication (2FA)** enabled, you are
   prompted for your authenticator code. See
   [Users & Access](13-users-and-access.md) for setting up 2FA.

> **Forgot your password?** Use **Forgot password** (`/forgot-password`), check
> your email, and set a new one on the **Reset password** screen.

## 1.4 First login — what happens automatically

On your first successful login, live-rack:

- Creates your user record inside your organization.
- Assigns the role granted to you (or a default low-privilege role).
- Maps you to the stores you are allowed to see.

If you cannot see any store or a screen says "forbidden," ask an **Admin** to
grant you a role or store access on the
[Users & Access](13-users-and-access.md) screen.

## 1.5 Choose your store

After login you land on the **Overview** dashboard. At the top of the screen is
the **store selector** (for example, "Store #14"). Click it to switch
warehouses. Every screen — map, inventory, tasks, orders — updates to the store
you pick.

## 1.6 Your role decides what you see

- **Admin / Manager** see operational controls (edit zones, create pipelines,
  manage users).
- **Staff** see day-to-day tools: scanner, inventory moves, their tasks,
  picking.
- **Read-only** see dashboards and can export, but cannot change data.

---

**Connects to:** [Navigation](02-navigation.md) ·
[Users & Access](13-users-and-access.md)
