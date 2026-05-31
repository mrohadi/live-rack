import type { Page } from "@playwright/test";

// AuthGuard derives org id from the project-roles claim inner key (LR-005a):
// { roleName: { orgId: orgDomain } }.
const ROLES_CLAIM = "urn:zitadel:iam:org:project:roles";

// seedOidcSession injects a ready react-oidc-context user into sessionStorage
// before the app boots, so AuthGuard renders authenticated without a UI login.
//
// E2E_OIDC_TOKEN must be a valid Zitadel access token for the test user.
// TODO LR-005a: mint this via a Zitadel service-account (JWT profile / PAT)
// in CI instead of reading a static env token.
export async function seedOidcSession(page: Page) {
  const issuer = process.env.VITE_OIDC_ISSUER ?? "http://localhost:8081";
  const clientId = process.env.VITE_OIDC_CLIENT_ID ?? "";
  // Non-empty fallback: useApi() throws on a falsy token before any fetch,
  // so stubbed-API specs need a dummy even without a real Zitadel token.
  const accessToken = process.env.E2E_OIDC_TOKEN || "e2e-access-token";
  const orgId = process.env.E2E_OIDC_ORG_ID ?? "00000000-0000-0000-0000-000000000001";

  const user = {
    access_token: accessToken,
    token_type: "Bearer",
    profile: {
      sub: process.env.E2E_OIDC_USER_ID ?? "e2e-user",
      name: "E2E User",
      email: "e2e@localhost",
      [ROLES_CLAIM]: { staff: { [orgId]: "localhost" } },
    },
    expires_at: Math.floor(Date.now() / 1000) + 3600,
  };

  await page.addInitScript(
    ([key, value]) => {
      window.sessionStorage.setItem(key, value);
    },
    [`oidc.user:${issuer}:${clientId}`, JSON.stringify(user)] as const,
  );
}
