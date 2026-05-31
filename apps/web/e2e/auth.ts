import type { Page } from "@playwright/test";

const ORG_CLAIM = "urn:zitadel:iam:user:resourceowner:id";

// seedOidcSession injects a ready react-oidc-context user into sessionStorage
// before the app boots, so AuthGuard renders authenticated without a UI login.
//
// E2E_OIDC_TOKEN must be a valid Zitadel access token for the test user.
// TODO LR-005a: mint this via a Zitadel service-account (JWT profile / PAT)
// in CI instead of reading a static env token.
export async function seedOidcSession(page: Page) {
  const issuer = process.env.VITE_OIDC_ISSUER ?? "http://localhost:8081";
  const clientId = process.env.VITE_OIDC_CLIENT_ID ?? "";
  const accessToken = process.env.E2E_OIDC_TOKEN ?? "";
  const orgId = process.env.E2E_OIDC_ORG_ID ?? "00000000-0000-0000-0000-000000000001";

  const user = {
    access_token: accessToken,
    token_type: "Bearer",
    profile: {
      sub: process.env.E2E_OIDC_USER_ID ?? "e2e-user",
      name: "E2E User",
      email: "e2e@localhost",
      [ORG_CLAIM]: orgId,
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
