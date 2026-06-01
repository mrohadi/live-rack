import { useAuth } from "react-oidc-context";
import { Outlet } from "react-router-dom";

import { Welcome } from "./auth/Welcome";
import { LoadingScreen } from "./LoadingScreen";

// Zitadel project-roles claim: { roleName: { orgId: orgDomain } }. Org id = inner key.
const ROLES_CLAIM = "urn:zitadel:iam:org:project:roles";

function orgIdFromRoles(profile?: Record<string, unknown>): string | undefined {
  const roles = profile?.[ROLES_CLAIM] as Record<string, Record<string, string>> | undefined;
  if (!roles) return undefined;
  const firstRole = Object.values(roles)[0];
  return firstRole ? Object.keys(firstRole)[0] : undefined;
}

export function AuthGuard() {
  const auth = useAuth();

  if (auth.isLoading || auth.activeNavigator) return <LoadingScreen />;

  // Signed out → branded landing. The user explicitly chooses to sign in
  // (hand-off to Zitadel) or create a workspace, instead of an abrupt redirect.
  if (!auth.isAuthenticated) {
    return (
      <Welcome
        onSignIn={() => {
          void auth.signinRedirect();
        }}
      />
    );
  }

  if (auth.error) {
    return (
      <div className="flex h-screen flex-col items-center justify-center gap-4">
        <p className="text-sm text-destructive">Sign-in failed: {auth.error.message}</p>
      </div>
    );
  }

  // Every user must belong to an org (tenant).
  const orgId = orgIdFromRoles(auth.user?.profile);
  if (!orgId) {
    return (
      <div className="flex h-screen items-center justify-center flex-col gap-4">
        <p className="text-sm text-muted-foreground">
          No organization found. Contact your administrator.
        </p>
      </div>
    );
  }

  return <Outlet />;
}
