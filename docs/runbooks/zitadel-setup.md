# Zitadel Setup & Deploy Checklist

Auth runs on self-hosted Zitadel (OIDC). This runbook covers everything
needed to bring auth up on a fresh server, plus how the onboarding token
(`ZITADEL_MGMT_TOKEN`) is created.

## 1. Run Zitadel

Pinned image (legacy hosted login):

```bash
docker compose -f deploy/docker/docker-compose.yml up -d zitadel zitadel-db
```

- Image: `ghcr.io/zitadel/zitadel:v2.71.9`
- Legacy login forced: `OIDC_DEFAULTLOGINURLV2=""`
- Console: `https://<host>/ui/console`
- First admin: `admin@localhost` / `Admin123!` (rotate on prod)

## 2. Project + SPA app

Create once per environment:

1. Console → **Projects → New** → `live-rack`. Record its **ID** → `OIDC_PROJECT_ID`.
2. Inside project → **Applications → New → User Agent (SPA), PKCE**.
   - Redirect URI: `https://<web-host>/callback`
   - Post-logout URI: `https://<web-host>/`
   - Record **Client ID** → `VITE_OIDC_CLIENT_ID`.
3. App → **Token Settings**:
   - Auth Token Type: **JWT**
   - ✅ Add user roles to access token
   - ✅ User roles inside ID token
4. Project → **Roles**: `admin`, `manager`, `staff`, `readonly`, `service`.

> Org id derives from the project-roles claim key
> `urn:zitadel:iam:org:project:{OIDC_PROJECT_ID}:roles` — no resourceowner claim.

## 3. Service user + ZITADEL_MGMT_TOKEN

The API uses a service-account PAT for onboarding (invite LR-907, signup
LR-908): create org, create human user, grant project role, reset password.

1. Org → **Users → Service Users → New** — e.g. `onboarding-bot`.
2. Service user → **Personal Access Tokens → New** → **copy the token now**
   (shown once). This value → `ZITADEL_MGMT_TOKEN`.
3. Grant authority:
   - Org → **Members → Add member** → service user → role **ORG_OWNER**
     (minimum: `ORG_USER_MANAGER`).
   - Project `live-rack` → **Authorizations → Add** → service user, so it can
     grant project roles.
4. Store the token in the deploy secret store (never commit). Inject as env
   `ZITADEL_MGMT_TOKEN`.

Verify the token:

```bash
curl -s -H "Authorization: Bearer $ZITADEL_MGMT_TOKEN" \
  https://<host>/management/v1/orgs/me | jq .org.id
```

Expect the org id. `401` → token wrong/expired. `403` → missing membership.

## 3b. SMTP (required for invite/verification emails)

Without SMTP, Zitadel generates the verification code but cannot deliver it.
Logs show `could not create email channel ... Errors.SMTPConfig.NotFound`, and
invitees never get the link.

We deliver via **Resend's SMTP relay** — keeps Zitadel's native invite flow and
templates, no app code. Run the helper (needs an **IAM_OWNER** PAT; SMTP config
is instance-scoped):

```bash
OIDC_ISSUER=https://<host> \
ZITADEL_ADMIN_TOKEN=<IAM_OWNER PAT> \
RESEND_API_KEY=re_xxx \
SMTP_SENDER_ADDRESS=no-reply@yourdomain.com \
SMTP_SENDER_NAME="live-rack" \
  scripts/zitadel-smtp.sh
```

Resend creds: host `smtp.resend.com:587`, user `resend`, password = `RESEND_API_KEY`.
`SMTP_SENDER_ADDRESS` must be a verified domain/address in Resend. The script
adds the provider and activates it.

Console equivalent: **Settings → Notifications → SMTP** → enter the same values →
**Test** → **Activate**.

No SMTP yet? Pull a code from logs to test manually:

```bash
docker logs docker-zitadel-1 2>&1 | grep -iE "Code:|verif" | tail -5
```

## 3c. Post-verification redirect to the app

After an invitee verifies their email / sets a password on the hosted pages,
Zitadel sends them to the login policy's **default redirect URI**. Unset → they
land on the Zitadel console instead of the app. Point it at the web app
(IAM_OWNER PAT, instance-scoped):

```bash
curl -s -X PUT "$OIDC_ISSUER/admin/v1/policies/login" \
  -H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" -H "Content-Type: application/json" \
  -d '{ "allowUsernamePassword": true, "allowRegister": true, "allowExternalIdp": true,
        "passwordlessType": "PASSWORDLESS_TYPE_ALLOWED", "allowDomainDiscovery": true,
        "secondFactors": ["SECOND_FACTOR_TYPE_OTP","SECOND_FACTOR_TYPE_U2F"],
        "multiFactors": ["MULTI_FACTOR_TYPE_U2F_WITH_VERIFICATION"],
        "defaultRedirectUri": "https://<web-host>" }'
```

Console equivalent: **Default settings → Login Behavior and Security → Default
Redirect URI**.

## 3d. Custom login UI (LR-909)

The app hosts its own sign-in at `/login` (and verify/reset screens) backed by
Zitadel's Session API via the `/api/v1/login/*` proxy. To route the SPA to it:

1. Console → Project `live-rack` → App `web` → **Login Behavior** →
   set **Login V2** with base URI `https://<web-host>` (origin only — Zitadel
   appends `/login?authRequest=…` itself). (Per-app override; console keeps legacy.)
2. The app's redirect URIs must include `https://<web-host>/callback` (already set)
   — `finalize` returns a callback URL there.
3. API env: `ZITADEL_LOGIN_CLIENT_TOKEN` (service user with **IAM_LOGIN_CLIENT**).
   Falls back to `ZITADEL_MGMT_TOKEN` if unset.
   **IAM_OWNER is NOT enough** — `finalize` (`/v2/oidc/auth_requests/{id}`)
   returns `No matching permissions (AUTH-AWfge)` without the **IAM_LOGIN_CLIENT**
   instance role. Grant it:
   `PUT /admin/v1/members/{userId}` body `{"roles":["IAM_OWNER","IAM_LOGIN_CLIENT"]}`.
4. Restart the API. Hit the app → it redirects to `/login?authRequest=…` → our UI.

Rollback: set the app's login back to **Login V1** in the console.

## 4. Branding (optional, matches palette)

```bash
ZITADEL_HOST=https://<host> \
ZITADEL_MGMT_TOKEN=$ZITADEL_MGMT_TOKEN \
  scripts/zitadel-branding.sh
```

Sets org LabelPolicy colors (primary `#2563eb`), then activates. Service user
must be ORG_OWNER on the target org. Restrictions/language set via Console.

## 5. Environment variables

API (`services/api`):

| Var | Example | Notes |
|---|---|---|
| `OIDC_ISSUER` | `https://<host>` | JWKS + token validation |
| `OIDC_PROJECT_ID` | `375294302823120905` | org-id claim key |
| `ZITADEL_MGMT_TOKEN` | `<PAT>` | onboarding service-account |

Web (`apps/web`):

| Var | Example |
|---|---|
| `VITE_OIDC_ISSUER` | `https://<host>` |
| `VITE_OIDC_CLIENT_ID` | `375295100143599625` |
| `VITE_OIDC_REDIRECT_URI` | `https://<web-host>/callback` |

After changing API env, **restart the API** so it reloads
`ZITADEL_MGMT_TOKEN`. An empty token → invite returns "Invite failed".

## 6. Smoke test

1. Sign in via hosted login → land on `/callback` → app loads.
2. Admin → Users & Access → **Add user** → invite a real email.
3. Recipient receives verification mail → sets password → enrols 2FA → signs in.
4. `403` on invite → caller not `admin`. "Invite failed" → check
   `ZITADEL_MGMT_TOKEN` present and SA membership.

## Common failures

| Symptom | Cause | Fix |
|---|---|---|
| "Invite failed" | `ZITADEL_MGMT_TOKEN` empty/unset | set env, restart API |
| `403` creating user | SA lacks ORG_OWNER | add org membership |
| `403` granting role | SA not in project authorizations | add project authorization |
| No roles in token | token settings off | enable roles in access + ID token |
| Login page unbranded | instance default, not org policy | run branding script as ORG_OWNER |
