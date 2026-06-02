import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "../../lib/api";

export interface OrgUser {
  id: string;
  idp_user_id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  role: string;
  title: string;
  shift: string;
  status: string;
  mfa_enabled: boolean;
  last_seen_at: string | null;
  zones: string[];
}

export interface AuditEntry {
  ts: string;
  actor_user_id: string | null;
  action: string;
  resource_type: string;
  resource_id: string;
  metadata: Record<string, unknown>;
}

export interface RosterStats {
  members: number;
  roles: number;
  active_now: number;
  pending_invites: number;
  twofa_coverage: number;
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
  stats: ["users", "stats"] as const,
};

/** Zone-access label: explicit zones joined, else org-wide "All". Pure. */
export function zonesLabel(zones: string[] | undefined): string {
  return zones && zones.length > 0 ? zones.join(", ") : "All";
}

/** Compact relative time from an ISO timestamp. Pure. */
export function relativeTime(iso: string | null, now: number = Date.now()): string {
  if (!iso) return "—";
  const diff = Math.max(0, now - new Date(iso).getTime());
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

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

/** Initials for an avatar from a display name or email. Pure.
 *  Emails fall back to their local part split on separators. */
export function initials(nameOrEmail: string): string {
  const raw = nameOrEmail.trim();
  if (!raw) return "?";
  const base = raw.includes("@") ? raw.slice(0, raw.indexOf("@")) : raw;
  const parts = base.split(/[\s._-]+/).filter(Boolean);
  const letters = parts
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
  return letters || "?";
}

/** Best display label for a user: their name, else email. Pure. */
export function displayLabel(u: { display_name: string; email: string }): string {
  return u.display_name.trim() || u.email;
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

/** Fetch the Users & Access header metrics. */
export function useRosterStats() {
  const { get } = useApi();
  return useQuery({
    queryKey: userKeys.stats,
    queryFn: () => get<RosterStats>("/api/v1/users/stats"),
  });
}

export interface TotpEnrollment {
  uri: string;
  secret: string;
}

/** Begin authenticator enrollment: returns the otpauth URI + manual secret. */
export function useStartTotp() {
  const { post } = useApi();
  return useMutation({
    mutationFn: () => post<TotpEnrollment>("/api/v1/me/2fa/totp", {}),
  });
}

/** Confirm authenticator enrollment with the first code; refreshes coverage. */
export function useVerifyTotp() {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (code: string) => post<void>("/api/v1/me/2fa/totp/verify", { code }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: userKeys.list });
      void qc.invalidateQueries({ queryKey: userKeys.stats });
    },
  });
}

/** Sync the caller's 2FA state (from the ID-token amr) to the server. */
export function useSyncMfa() {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (enabled: boolean) => post<void>("/api/v1/me/2fa", { enabled }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: userKeys.list });
      void qc.invalidateQueries({ queryKey: userKeys.stats });
    },
  });
}

/** Roles an admin may assign when inviting a teammate. */
export const ASSIGNABLE_ROLES = ["admin", "manager", "staff", "readonly"] as const;
export type AssignableRole = (typeof ASSIGNABLE_ROLES)[number];

export interface InvitePayload {
  email: string;
  display_name: string;
  role: AssignableRole;
}

export interface InviteResult {
  user_id: string;
  email: string;
  role: string;
  status: string;
}

// mfaMethods mark a second factor in an OIDC amr claim.
const mfaMethods = new Set(["mfa", "otp", "totp", "webauthn", "u2f", "hwk"]);

/** True when the OIDC profile's amr claim shows a second factor. Pure.
 *  amr lives in the ID token (profile), not the access token. */
export function hasMfa(profile: Record<string, unknown> | undefined): boolean {
  const amr = profile?.amr;
  return Array.isArray(amr) && amr.some((m) => mfaMethods.has(String(m).toLowerCase()));
}

/** True when the caller may invite users. Authorizes on the admin role; MFA is
 *  enforced at the Zitadel login policy (the access token carries no amr). Pure. */
export function canInvite(caps: Capabilities | undefined): boolean {
  return Boolean(caps && caps.role === "admin");
}

/** Invite a teammate; refreshes the roster on success. */
export function useInviteUser() {
  const { post } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: InvitePayload) => post<InviteResult>("/api/v1/users/invite", body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: userKeys.list }),
  });
}

/** Resend a pending invite by Zitadel user id. */
export function useResendInvite() {
  const { post } = useApi();
  return useMutation({
    mutationFn: (userID: string) => post<void>(`/api/v1/users/${userID}/resend`, {}),
  });
}

/** Recent audit-trail entries, optionally scoped to one actor. */
export function useAudit(actor?: string, limit = 10) {
  const { get } = useApi();
  const qs = `?limit=${limit}${actor ? `&actor=${actor}` : ""}`;
  return useQuery({
    queryKey: ["users", "audit", actor ?? "all", limit] as const,
    queryFn: () => get<AuditEntry[]>(`/api/v1/audit${qs}`),
  });
}

/** Change a member's role (re-grants in Zitadel via idp id). */
export function useSetRole() {
  const { patch } = useApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (v: { id: string; idpUserId: string; role: AssignableRole }) =>
      patch<void>(`/api/v1/users/${v.id}/role`, { role: v.role, idp_user_id: v.idpUserId }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: userKeys.list }),
  });
}

/** Email a member a password-reset link (by Zitadel user id). */
export function useResetPassword() {
  const { post } = useApi();
  return useMutation({
    mutationFn: (idpUserId: string) => post<void>(`/api/v1/users/${idpUserId}/reset-password`, {}),
  });
}

/** Humanize an audit action key, e.g. "user.role_changed" → "Role changed". Pure. */
export function auditLabel(action: string): string {
  const tail = action.includes(".") ? action.split(".").slice(1).join(".") : action;
  const s = tail.replace(/_/g, " ");
  return s.charAt(0).toUpperCase() + s.slice(1);
}
