const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export interface LoginSession {
  session_id: string;
  session_token: string;
}

export interface StartResult extends LoginSession {
  mfa_required: boolean;
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE_URL}/api/v1/login/${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`${res.status}: ${text}`);
  }
  return res.json() as Promise<T>;
}

/** Open a session for a login name (email); reports whether MFA is required. */
export function startSession(loginName: string): Promise<StartResult> {
  return post<StartResult>("start", { login_name: loginName });
}

/** Verify a password against an open session. */
export function checkPassword(s: LoginSession, password: string): Promise<LoginSession> {
  return post<LoginSession>("password", {
    session_id: s.session_id,
    session_token: s.session_token,
    password,
  });
}

/** Verify a TOTP authenticator code against an open session. */
export function checkTotp(s: LoginSession, code: string): Promise<LoginSession> {
  return post<LoginSession>("totp", {
    session_id: s.session_id,
    session_token: s.session_token,
    code,
  });
}

/** Bind a verified session to the OIDC auth request; returns the callback URL. */
export function finalize(
  authRequestId: string,
  s: LoginSession,
): Promise<{ callback_url: string }> {
  return post<{ callback_url: string }>("finalize", {
    auth_request_id: authRequestId,
    session_id: s.session_id,
    session_token: s.session_token,
  });
}

/** Extract a short, human-readable reason from a thrown login error. Pure. */
export function loginErrorMessage(err: unknown): string {
  const raw = err instanceof Error ? err.message : String(err);
  if (raw.startsWith("401")) return "Incorrect email or password.";
  if (raw.startsWith("400")) return "Please check the details and try again.";
  return "Something went wrong. Try again.";
}
