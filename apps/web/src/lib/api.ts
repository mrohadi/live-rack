import { useAuth } from "react-oidc-context";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

async function request<T>(path: string, token: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.text().catch(() => "");
    throw new Error(`${res.status} ${res.statusText}: ${body}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

// useApi returns a typed fetch helper pre-loaded with the OIDC access token.
export function useApi() {
  const auth = useAuth();
  const getToken = () => auth.user?.access_token ?? null;

  return {
    get: async <T>(path: string): Promise<T> => {
      const token = getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token);
    },
    post: async <T>(path: string, body: unknown): Promise<T> => {
      const token = getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token, { method: "POST", body: JSON.stringify(body) });
    },
    put: async <T>(path: string, body: unknown): Promise<T> => {
      const token = getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token, { method: "PUT", body: JSON.stringify(body) });
    },
    patch: async <T>(path: string, body: unknown): Promise<T> => {
      const token = getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token, { method: "PATCH", body: JSON.stringify(body) });
    },
    del: async <T>(path: string): Promise<T> => {
      const token = getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token, { method: "DELETE" });
    },
  };
}
