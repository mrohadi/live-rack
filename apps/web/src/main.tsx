import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AuthProvider, type AuthProviderProps } from "react-oidc-context";
import { RouterProvider, createBrowserRouter } from "react-router-dom";

import "./styles/index.css";
import { routes } from "./routes";
import { ToastProvider } from "./components/feedback/toast";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
});

const issuer = import.meta.env.VITE_OIDC_ISSUER;
const clientId = import.meta.env.VITE_OIDC_CLIENT_ID;
if (!issuer || !clientId) throw new Error("Missing VITE_OIDC_ISSUER or VITE_OIDC_CLIENT_ID");

const oidcConfig: AuthProviderProps = {
  authority: issuer,
  client_id: clientId,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI ?? `${window.location.origin}/callback`,
  post_logout_redirect_uri: window.location.origin,
  // openid+profile+email for identity; the roles scope emits the project-roles claim.
  scope: "openid profile email urn:zitadel:iam:org:project:roles",
  response_type: "code",
  // Merge userinfo claims into profile — Zitadel emits resourceowner/org id there.
  loadUserInfo: true,
  // Strip the ?code/&state from the URL after a successful sign-in.
  onSigninCallback: () => {
    window.history.replaceState({}, document.title, "/");
  },
};

const router = createBrowserRouter(routes);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AuthProvider {...oidcConfig}>
      <QueryClientProvider client={queryClient}>
        <ToastProvider>
          <RouterProvider router={router} />
        </ToastProvider>
      </QueryClientProvider>
    </AuthProvider>
  </StrictMode>,
);
