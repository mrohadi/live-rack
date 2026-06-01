import { useAuth } from "react-oidc-context";
import { Navigate, Outlet } from "react-router-dom";
import { isAdmin } from "../lib/roles";

/** Route guard: renders child routes only for admins; others bounce to home. */
export function RequireAdmin() {
  const auth = useAuth();
  return isAdmin(auth.user?.profile) ? <Outlet /> : <Navigate to="/" replace />;
}
