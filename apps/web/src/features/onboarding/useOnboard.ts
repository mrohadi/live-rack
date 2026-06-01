const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export interface CompletePayload {
  user_id: string;
  org_id: string;
  code: string;
  password: string;
}

/** Accept an invite: verify the email + set the password. */
export async function completeOnboarding(body: CompletePayload): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/v1/onboard/complete`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`${res.status}: ${text}`);
  }
}

export interface TotpEnrollment {
  uri: string;
  secret: string;
}

/** Begin authenticator enrollment during onboarding (gated by the new password). */
export async function startOnboardTotp(userID: string, password: string): Promise<TotpEnrollment> {
  const res = await fetch(`${BASE_URL}/api/v1/onboard/totp/start`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userID, password }),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`${res.status}: ${text}`);
  }
  return res.json() as Promise<TotpEnrollment>;
}

/** Confirm authenticator enrollment with the first code (gated by the password). */
export async function verifyOnboardTotp(
  userID: string,
  password: string,
  code: string,
): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/v1/onboard/totp/verify`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userID, password, code }),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`${res.status}: ${text}`);
  }
}

export interface PasswordRule {
  label: string;
  ok: boolean;
}

/** Evaluate a password (and confirmation) against the Zitadel default policy.
 *  Pure — returns a rule list the UI renders as a checklist. */
export function passwordRules(password: string, confirm: string): PasswordRule[] {
  return [
    { label: "At least 8 characters", ok: password.length >= 8 },
    { label: "An uppercase letter", ok: /[A-Z]/.test(password) },
    { label: "A lowercase letter", ok: /[a-z]/.test(password) },
    { label: "A number", ok: /[0-9]/.test(password) },
    { label: "A symbol", ok: /[^A-Za-z0-9]/.test(password) },
    { label: "Passwords match", ok: password.length > 0 && password === confirm },
  ];
}

/** True when every password rule passes. Pure. */
export function passwordValid(password: string, confirm: string): boolean {
  return passwordRules(password, confirm).every((r) => r.ok);
}

/** Short, human-readable reason from a thrown onboarding error. Pure. */
export function onboardErrorMessage(err: unknown): string {
  const raw = err instanceof Error ? err.message : String(err);
  if (raw.includes("invalid or expired")) return "This link is invalid or has expired.";
  if (raw.startsWith("400")) return "Please check the details and try again.";
  return "Something went wrong. Try again.";
}
