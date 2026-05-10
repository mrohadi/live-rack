import { useAuth, useOrganization } from "@clerk/clerk-react";
import { Navigate, Outlet } from "react-router-dom";

export function AuthGuard() {
  const { isLoaded, isSignedIn } = useAuth();
  const { isLoaded: orgLoaded, organization } = useOrganization();

  if (!isLoaded || !orgLoaded) {
    return (
      <div className="flex h-screen items-center justify-center">
        <span className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
      </div>
    );
  }

  if (!isSignedIn) return <Navigate to="/sign-in" replace />;

  // Clerk org required — every user must belong to an org (tenant).
  if (!organization) {
    return (
      <div className="flex h-screen items-center justify-center flex-col gap-4">
        <p className="text-sm text-muted-foreground">No organization found. Contact your administrator.</p>
      </div>
    );
  }

  return <Outlet />;
}
