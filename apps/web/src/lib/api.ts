import { useAuth } from "@clerk/clerk-react";

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

// useApi returns a typed fetch helper pre-loaded with the Clerk JWT.
export function useApi() {
  const { getToken } = useAuth();

  return {
    get: async <T>(path: string): Promise<T> => {
      const token = await getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token);
    },
    post: async <T>(path: string, body: unknown): Promise<T> => {
      const token = await getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token, { method: "POST", body: JSON.stringify(body) });
    },
    put: async <T>(path: string, body: unknown): Promise<T> => {
      const token = await getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token, { method: "PUT", body: JSON.stringify(body) });
    },
    del: async <T>(path: string): Promise<T> => {
      const token = await getToken();
      if (!token) throw new Error("Not authenticated");
      return request<T>(path, token, { method: "DELETE" });
    },
  };
}
