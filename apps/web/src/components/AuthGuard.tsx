import { useEffect } from "react";
import { useAuth } from "react-oidc-context";
import { Outlet } from "react-router-dom";

// Zitadel project-roles claim: { roleName: { orgId: orgDomain } }. Org id = inner key.
const ROLES_CLAIM = "urn:zitadel:iam:org:project:roles";

function orgIdFromRoles(profile?: Record<string, unknown>): string | undefined {
  const roles = profile?.[ROLES_CLAIM] as Record<string, Record<string, string>> | undefined;
  if (!roles) return undefined;
  const firstRole = Object.values(roles)[0];
  return firstRole ? Object.keys(firstRole)[0] : undefined;
}

function Spinner({ label }: { label: string }) {
  return (
    <div
      style={{
        display: "flex",
        height: "100vh",
        alignItems: "center",
        justifyContent: "center",
        color: "#94a3b8",
      }}
    >
      {label}
    </div>
  );
}

export function AuthGuard() {
  const auth = useAuth();

  // Kick off the redirect to Zitadel's hosted login once we know the user is signed out.
  useEffect(() => {
    if (!auth.isLoading && !auth.isAuthenticated && !auth.activeNavigator && !auth.error) {
      void auth.signinRedirect();
    }
  }, [auth]);

  if (auth.isLoading || auth.activeNavigator)
    return <Spinner label={`Loading… (${auth.activeNavigator ?? "init"})`} />;

  if (auth.error) {
    return (
      <div className="flex h-screen items-center justify-center flex-col gap-4">
        <p className="text-sm text-red-500">Sign-in failed: {auth.error.message}</p>
      </div>
    );
  }

  if (!auth.isAuthenticated) return <Spinner label="Redirecting to sign-in…" />;

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
