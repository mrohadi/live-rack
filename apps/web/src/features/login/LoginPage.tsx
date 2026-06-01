import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { AuthLayout } from "../../components/auth/AuthLayout";
import {
  checkPassword,
  checkTotp,
  finalize,
  loginErrorMessage,
  startSession,
  type LoginSession,
} from "./useLogin";

const field =
  "w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary";
const primaryBtn =
  "w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-white disabled:opacity-50";

/** Custom sign-in UI backed by Zitadel's Session API (via our login proxy).
 *  Email + password on one page; a TOTP step appears only when the account has a
 *  second factor. The OIDC auth request id arrives as ?authRequest= from Zitadel. */
const AUTH_REQ_KEY = "lr.authRequest";

export function LoginPage() {
  const [params] = useSearchParams();
  // The auth request id arrives once in the URL. Persist it so it survives the
  // TOTP step and any refresh; finalize needs it.
  const [authRequestId, setAuthRequestId] = useState(
    () => params.get("authRequest") ?? sessionStorage.getItem(AUTH_REQ_KEY) ?? "",
  );

  useEffect(() => {
    const fromUrl = params.get("authRequest");
    if (fromUrl) {
      sessionStorage.setItem(AUTH_REQ_KEY, fromUrl);
      setAuthRequestId(fromUrl);
    }
  }, [params]);

  // /login is only meaningful with an auth request from Zitadel. If opened
  // directly, bounce to the app so the OIDC flow issues one.
  useEffect(() => {
    if (!authRequestId) {
      const t = setTimeout(() => window.location.replace("/"), 1200);
      return () => clearTimeout(t);
    }
  }, [authRequestId]);

  const [credsDone, setCredsDone] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [session, setSession] = useState<LoginSession | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const goToApp = async (s: LoginSession) => {
    const { callback_url } = await finalize(authRequestId, s);
    sessionStorage.removeItem(AUTH_REQ_KEY);
    window.location.href = callback_url;
  };

  // Email + password: one submit. Branches to TOTP only when MFA is configured.
  const submitCreds = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const started = await startSession(email.trim());
      const s = await checkPassword(started, password);
      setSession(s);
      if (started.mfa_required) {
        setCredsDone(true); // show TOTP step
      } else {
        await goToApp(s);
      }
    } catch (err) {
      setError(loginErrorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  const submitTotp = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!session) return;
    setBusy(true);
    setError("");
    try {
      const s = await checkTotp(session, code.trim());
      await goToApp(s);
    } catch (err) {
      setError(loginErrorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthLayout
      title="Sign in"
      subtitle={
        credsDone
          ? "Enter the 6-digit code from your authenticator app."
          : "Welcome back. Sign in to your live-rack workspace."
      }
    >
      {!authRequestId ? (
        <p role="status" className="rounded-md bg-warning/10 p-3 text-xs text-warning">
          Redirecting to live-rack to start sign-in…
        </p>
      ) : !credsDone ? (
        <form onSubmit={submitCreds} className="space-y-4">
          <label className="block text-sm">
            <span className="mb-1 block text-muted-foreground">Email</span>
            <input
              type="email"
              required
              autoFocus
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@company.com"
              className={field}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-muted-foreground">Password</span>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className={field}
            />
          </label>
          {error && (
            <p role="alert" className="text-xs text-destructive">
              {error}
            </p>
          )}
          <button type="submit" disabled={busy} className={primaryBtn}>
            {busy ? "Signing in…" : "Sign in"}
          </button>
          <Link to="/forgot-password" className="block text-center text-xs text-primary">
            Forgot password?
          </Link>
        </form>
      ) : (
        <form onSubmit={submitTotp} className="space-y-4">
          <label className="block text-sm">
            <span className="mb-1 block text-muted-foreground">Authentication code</span>
            <input
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              maxLength={6}
              required
              autoFocus
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="123456"
              className={`${field} tracking-[0.5em]`}
            />
          </label>
          {error && (
            <p role="alert" className="text-xs text-destructive">
              {error}
            </p>
          )}
          <button type="submit" disabled={busy} className={primaryBtn}>
            {busy ? "Verifying…" : "Verify"}
          </button>
        </form>
      )}

      <p className="mt-6 text-center text-sm text-muted-foreground">
        New to live-rack?{" "}
        <Link to="/signup" className="font-medium text-primary">
          Create an account
        </Link>
      </p>
    </AuthLayout>
  );
}
