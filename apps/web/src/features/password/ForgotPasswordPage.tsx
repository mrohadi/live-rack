import { useState } from "react";
import { Link } from "react-router-dom";
import { AuthLayout } from "../../components/auth/AuthLayout";
import { requestPasswordReset, resetErrorMessage } from "./usePasswordReset";

const field =
  "w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary";
const primaryBtn =
  "w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-white disabled:opacity-50";

/** Request a password-reset link. The response never reveals whether the email
 *  is registered; we always show the same confirmation. */
export function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [busy, setBusy] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await requestPasswordReset(email.trim().toLowerCase());
      setSent(true);
    } catch (err) {
      setError(resetErrorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  if (sent) {
    return (
      <AuthLayout
        title="Check your email"
        subtitle="If that account exists, a reset link is on its way."
      >
        <p role="status" className="text-sm text-muted-foreground">
          We sent a password-reset link to <strong>{email}</strong>. The link expires soon.
        </p>
        <p className="mt-6 text-center text-sm text-muted-foreground">
          <Link to="/login" className="font-medium text-primary">
            Back to sign in
          </Link>
        </p>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      title="Forgot password"
      subtitle="Enter your email and we'll send a link to reset your password."
    >
      <form onSubmit={submit} className="space-y-4">
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
        {error && (
          <p role="alert" className="text-xs text-destructive">
            {error}
          </p>
        )}
        <button type="submit" disabled={busy} className={primaryBtn}>
          {busy ? "Sending…" : "Send reset link"}
        </button>
        <Link to="/login" className="block text-center text-xs text-primary">
          Back to sign in
        </Link>
      </form>
    </AuthLayout>
  );
}
