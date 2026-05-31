import { useEffect } from "react";
import { useAuth } from "react-oidc-context";
import { useNavigate } from "react-router-dom";

import { LoadingScreen } from "./LoadingScreen";

// AuthProvider processes the OIDC redirect automatically. onSigninCallback rewrites
// the URL to "/" via raw history API, which react-router never observes — so once
// auth settles we navigate through the router to leave this spinner.
export function CallbackPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!auth.isLoading && !auth.activeNavigator) {
      navigate("/", { replace: true });
    }
  }, [auth.isLoading, auth.activeNavigator, navigate]);

  return <LoadingScreen label="Signing in…" />;
}
