import { useMutation } from "@tanstack/react-query";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export interface SignupPayload {
  company: string;
  email: string;
  display_name: string;
}

export interface SignupResult {
  org_id: string;
  user_id: string;
  status: string;
  verify_url?: string;
}

/** Self-service signup — public, no auth token. Provisions a tenant + admin. */
export function useSignup() {
  return useMutation({
    mutationFn: async (body: SignupPayload): Promise<SignupResult> => {
      const res = await fetch(`${BASE_URL}/api/v1/signup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        if (res.status === 409) throw new Error("email_taken");
        const text = await res.text().catch(() => "");
        throw new Error(`${res.status}: ${text}`);
      }
      return res.json() as Promise<SignupResult>;
    },
  });
}
