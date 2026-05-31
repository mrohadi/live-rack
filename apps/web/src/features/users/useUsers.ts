import { useQuery } from "@tanstack/react-query";
import { useApi } from "../../lib/api";

export interface OrgUser {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  role: string;
}

export interface Capabilities {
  user_id: string;
  role: string;
  mfa_verified: boolean;
  permissions: string[];
  store_scoped: boolean;
  zone_scoped: boolean;
}

export const userKeys = {
  list: ["users", "list"] as const,
  me: ["users", "me"] as const,
};

/** Role columns shown in the permission matrix, in display order. */
export const ROLE_COLUMNS = ["Admin", "Manager", "Staff", "Read-only"] as const;

/** Static role × permission matrix mirroring the design grid. */
export const PERMISSION_MATRIX: { label: string; allow: boolean[] }[] = [
  { label: "View dashboards", allow: [true, true, true, true] },
  { label: "Edit zones & layout", allow: [true, true, false, false] },
  { label: "Approve mis-scans", allow: [true, true, false, false] },
  { label: "Manage pipelines", allow: [true, true, false, false] },
  { label: "Run scanner", allow: [true, true, true, false] },
  { label: "Move inventory", allow: [true, true, true, false] },
  { label: "Manage tasks (any)", allow: [true, true, false, false] },
  { label: "Manage tasks (own)", allow: [true, true, true, false] },
  { label: "Edit users", allow: [true, false, false, false] },
  { label: "Manage integrations", allow: [true, false, false, false] },
  { label: "Export reports", allow: [true, true, false, true] },
];

/** Initials for an avatar from a display name. Pure. */
export function initials(name: string): string {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
}

/** Fetch the org user roster. */
export function useUsers() {
  const { get } = useApi();
  return useQuery({ queryKey: userKeys.list, queryFn: () => get<OrgUser[]>("/api/v1/users") });
}

/** Fetch the caller's capabilities. */
export function useCapabilities() {
  const { get } = useApi();
  return useQuery({ queryKey: userKeys.me, queryFn: () => get<Capabilities>("/api/v1/me") });
}
