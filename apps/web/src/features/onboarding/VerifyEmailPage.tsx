import { useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { AuthLayout } from "../../components/auth/AuthLayout";
import {
  completeOnboarding,
  onboardErrorMessage,
  passwordRules,
  passwordValid,
  startOnboardTotp,
  verifyOnboardTotp,
  type TotpEnrollment,
} from "./useOnboard";

const field =
  "w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary";
const primaryBtn =
  "w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-white disabled:opacity-50";

type Step = "password" | "totp" | "done";

/** Custom invite-acceptance screen. The invite email links here with code,
 *  userID, and orgID. The user sets a password, enrolls an authenticator, then
 *  signs in. Replaces the Zitadel hosted verification page. */
export function VerifyEmailPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();

  const code = params.get("code") ?? "";
  const userID = params.get("userID") ?? "";
  const orgID = params.get("orgID") ?? "";
  const linkOk = Boolean(code && userID && orgID);

  const [step, setStep] = useState<Step>("password");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [totp, setTotp] = useState<TotpEnrollment | null>(null);
  const [otp, setOtp] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const rules = passwordRules(password, confirm);
  const valid = passwordValid(password, confirm);

  const finish = () => {
    setStep("done");
    setTimeout(() => navigate("/login"), 1500);
  };

  // Set password, then begin authenticator enrollment.
  const submitPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!valid || !linkOk) return;
    setBusy(true);
    setError("");
    try {
      await completeOnboarding({ user_id: userID, org_id: orgID, code, password });
      const enrollment = await startOnboardTotp(userID, password);
      setTotp(enrollment);
      setStep("totp");
    } catch (err) {
      setError(onboardErrorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  const submitOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await verifyOnboardTotp(userID, password, otp.trim());
      finish();
    } catch (err) {
      setError(onboardErrorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  if (!linkOk) {
    return (
      <AuthLayout title="Verify your email" subtitle="This invitation link is incomplete.">
        <p role="alert" className="rounded-md bg-destructive/10 p-3 text-xs text-destructive">
          The link is missing required details. Ask your admin to resend the invite.
        </p>
      </AuthLayout>
    );
  }

  if (step === "done") {
    return (
      <AuthLayout title="You're all set" subtitle="Your account is ready.">
        <p role="status" className="text-sm text-muted-foreground">
          Email verified, password set, authenticator enrolled. Taking you to sign in…
        </p>
      </AuthLayout>
    );
  }

  if (step === "totp") {
    return (
      <AuthLayout
        title="Set up authenticator"
        subtitle="Scan the code with your authenticator app, then enter the 6-digit code."
      >
        <div className="space-y-4">
          <div className="flex justify-center rounded-lg border border-border bg-white p-4">
            {totp ? (
              <QRCodeSVG value={totp.uri} size={176} />
            ) : (
              <span className="text-sm text-muted-foreground">Loading…</span>
            )}
          </div>
          {totp && (
            <p className="text-center text-[11px] text-muted-foreground">
              Can&apos;t scan? Enter this key manually:{" "}
              <span className="font-mono text-foreground">{totp.secret}</span>
            </p>
          )}
          <form onSubmit={submitOtp} className="space-y-4">
            <label className="block text-sm">
              <span className="mb-1 block text-muted-foreground">6-digit code</span>
              <input
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                required
                autoFocus
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/\D/g, ""))}
                placeholder="123456"
                className={`${field} text-center tracking-[0.5em]`}
              />
            </label>
            {error && (
              <p role="alert" className="text-xs text-destructive">
                {error}
              </p>
            )}
            <button type="submit" disabled={busy || otp.length !== 6} className={primaryBtn}>
              {busy ? "Verifying…" : "Verify & finish"}
            </button>
            <button
              type="button"
              onClick={finish}
              className="block w-full text-center text-xs text-muted-foreground hover:text-foreground"
            >
              Skip for now
            </button>
          </form>
        </div>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      title="Verify your email"
      subtitle="Confirm your address and choose a password to finish setting up your account."
    >
      <form onSubmit={submitPassword} className="space-y-4">
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
          {busy ? "Finishing…" : "Continue"}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Already verified?{" "}
        <Link to="/login" className="font-medium text-primary">
          Sign in
        </Link>
      </p>
    </AuthLayout>
  );
}
