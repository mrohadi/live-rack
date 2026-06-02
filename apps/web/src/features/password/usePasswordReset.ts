const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

/** Request a password-reset email. Always resolves (no user enumeration). */
export async function requestPasswordReset(email: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/v1/password/forgot`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`${res.status}: ${text}`);
  }
}

/** Set a new password using a reset code. */
export async function resetPassword(userID: string, code: string, password: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/v1/password/reset`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userID, code, password }),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`${res.status}: ${text}`);
  }
}

/** Short, human-readable reason from a thrown reset error. Pure. */
export function resetErrorMessage(err: unknown): string {
  const raw = err instanceof Error ? err.message : String(err);
  if (raw.includes("invalid or expired")) return "This link is invalid or has expired.";
  if (raw.startsWith("400")) return "Please check the details and try again.";
  return "Something went wrong. Try again.";
}
