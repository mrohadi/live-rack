import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { AuthLayout } from "../../components/auth/AuthLayout";
import { passwordRules, passwordValid } from "../onboarding/useOnboard";
import { resetErrorMessage, resetPassword } from "./usePasswordReset";

const field =
  "w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary";
const primaryBtn =
  "w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-white disabled:opacity-50";

/** Set a new password from a reset link. The email links here with code +
 *  userID; the user chooses a policy-compliant password. */
export function ResetPasswordPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();

  const code = params.get("code") ?? "";
  const userID = params.get("userID") ?? "";
  const linkOk = Boolean(code && userID);

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [done, setDone] = useState(false);

  const rules = passwordRules(password, confirm);
  const valid = passwordValid(password, confirm);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!valid || !linkOk) return;
    setBusy(true);
    setError("");
    try {
      await resetPassword(userID, code, password);
      setDone(true);
      setTimeout(() => navigate("/login"), 1500);
    } catch (err) {
      setError(resetErrorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  if (!linkOk) {
    return (
      <AuthLayout title="Reset password" subtitle="This reset link is incomplete.">
        <p role="alert" className="rounded-md bg-destructive/10 p-3 text-xs text-destructive">
          The link is missing required details. Request a new one.
        </p>
        <p className="mt-6 text-center text-sm text-muted-foreground">
          <Link to="/forgot-password" className="font-medium text-primary">
            Forgot password
          </Link>
        </p>
      </AuthLayout>
    );
  }

  if (done) {
    return (
      <AuthLayout title="Password updated" subtitle="You can now sign in.">
        <p role="status" className="text-sm text-muted-foreground">
          Your password has been reset. Taking you to sign in…
        </p>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout title="Reset password" subtitle="Choose a new password for your account.">
      <form onSubmit={submit} className="space-y-4">
        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">New password</span>
          <input
            type="password"
            required
            autoFocus
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            className={field}
          />
        </label>
        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Confirm password</span>
          <input
            type="password"
            required
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            placeholder="••••••••"
            className={field}
          />
        </label>

        <ul className="grid grid-cols-2 gap-x-3 gap-y-1 text-[11px]">
          {rules.map((r) => (
            <li
              key={r.label}
              className={`flex items-center gap-1.5 ${r.ok ? "text-success" : "text-muted-foreground"}`}
            >
              <span>{r.ok ? "✓" : "○"}</span>
              {r.label}
            </li>
          ))}
        </ul>

        {error && (
          <p role="alert" className="text-xs text-destructive">
            {error}
          </p>
        )}

        <button type="submit" disabled={busy || !valid} className={primaryBtn}>
          {busy ? "Updating…" : "Update password"}
        </button>
      </form>
    </AuthLayout>
  );
}
