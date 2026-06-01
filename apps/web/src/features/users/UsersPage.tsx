import { useEffect, useRef, useState } from "react";
import { useAuth } from "react-oidc-context";
import { AuditLogModal } from "./AuditLogModal";
import { EditAccessModal } from "./EditAccessModal";
import { Enroll2FAModal } from "./Enroll2FAModal";
import { InviteUserModal } from "./InviteUserModal";
import {
  PERMISSION_MATRIX,
  ROLE_COLUMNS,
  auditLabel,
  canInvite,
  displayLabel,
  hasMfa,
  initials,
  relativeTime,
  useAudit,
  useCapabilities,
  useResendInvite,
  useResetPassword,
  useRosterStats,
  useSyncMfa,
  useUsers,
  zonesLabel,
  type OrgUser,
} from "./useUsers";

const ROLE_TONES: Record<string, string> = {
  admin: "bg-destructive/15 text-destructive",
  manager: "bg-primary/15 text-primary",
  staff: "bg-success/15 text-success",
  readonly: "bg-muted text-muted-foreground",
  service: "bg-violet-500/15 text-violet-500",
};

// Deterministic per-member avatar color, like the design roster.
const AVATAR_TONES = [
  "bg-primary",
  "bg-success",
  "bg-violet-500",
  "bg-amber-500",
  "bg-rose-500",
  "bg-cyan-500",
  "bg-teal-500",
  "bg-indigo-500",
];

function avatarTone(seed: string): string {
  let h = 0;
  for (const c of seed) h = (h * 31 + c.charCodeAt(0)) >>> 0;
  return AVATAR_TONES[h % AVATAR_TONES.length];
}

function RoleChip({ role }: { role: string }) {
  return (
    <span
      className={`rounded px-1.5 py-0.5 text-[11px] font-medium capitalize ${ROLE_TONES[role] ?? "bg-muted text-muted-foreground"}`}
    >
      {role || "—"}
    </span>
  );
}

function StatusChip({ status }: { status: string }) {
  const map: Record<string, { c: string; label: string }> = {
    active: { c: "bg-success/15 text-success", label: "Active" },
    break: { c: "bg-warning/15 text-warning", label: "On break" },
    off: { c: "bg-muted text-muted-foreground", label: "Off" },
    pending: { c: "bg-warning/15 text-warning", label: "Invited" },
  };
  const s = map[status] ?? map.off;
  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs ${s.c}`}>
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {s.label}
    </span>
  );
}

function StatCard({ label, value, meta }: { label: string; value: string; meta?: string }) {
  return (
    <div className="rounded-lg border border-border bg-surface p-4">
      <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </div>
      <div className="mt-1 text-2xl font-semibold text-foreground">{value}</div>
      {meta && <div className="mt-1 text-xs text-muted-foreground">{meta}</div>}
    </div>
  );
}

const ROLE_TABS: { key: string; label: string }[] = [
  { key: "all", label: "All" },
  { key: "admin", label: "Admin" },
  { key: "manager", label: "Manager" },
  { key: "staff", label: "Staff" },
  { key: "readonly", label: "Read-only" },
  { key: "service", label: "Service" },
];

function Avatar({ name, size = 32 }: { name: string; size?: number }) {
  return (
    <div
      className={`flex shrink-0 items-center justify-center rounded-full font-semibold text-white ${avatarTone(name)}`}
      style={{ width: size, height: size, fontSize: size > 36 ? 14 : 11 }}
    >
      {initials(name)}
    </div>
  );
}

export function UsersPage() {
  const usersQuery = useUsers();
  const me = useCapabilities();
  const stats = useRosterStats();
  const auth = useAuth();
  const syncMfa = useSyncMfa();

  const users = usersQuery.data ?? [];
  const [filter, setFilter] = useState("all");
  const [selectedId, setSelectedId] = useState<string | undefined>();
  const [inviting, setInviting] = useState(false);
  const [auditOpen, setAuditOpen] = useState(false);
  const [editing, setEditing] = useState(false);
  const [enrolling, setEnrolling] = useState(false);
  const [linkCopied, setLinkCopied] = useState(false);
  const resetPassword = useResetPassword();
  const resendInvite = useResendInvite();

  const showInvite = canInvite(me.data);
  const mfaOn = hasMfa(auth.user?.profile);

  const copyInviteLink = () => {
    void navigator.clipboard?.writeText(`${window.location.origin}/signup`);
    setLinkCopied(true);
    setTimeout(() => setLinkCopied(false), 1500);
  };

  // Sync the caller's 2FA state (amr lives only in the ID token) once it is known.
  const synced = useRef(false);
  useEffect(() => {
    if (mfaOn && me.data && !synced.current) {
      synced.current = true;
      syncMfa.mutate(true);
    }
  }, [mfaOn, me.data, syncMfa]);

  const filtered = users.filter((u) => filter === "all" || u.role === filter);
  const selected: OrgUser | undefined =
    users.find((u) => u.id === selectedId) ?? filtered[0] ?? users[0];

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <h1 className="text-lg font-semibold text-foreground">Users &amp; Access</h1>
          <p className="text-xs text-muted-foreground">
            {stats.data?.members ?? users.length} members · {stats.data?.roles ?? "—"} roles
            {me.data ? ` · you: ${me.data.role}${mfaOn ? " · 2FA on" : ""}` : ""}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setAuditOpen(true)}
            className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition hover:bg-muted"
          >
            Audit log
          </button>
          {showInvite && (
            <>
              <button
                type="button"
                onClick={copyInviteLink}
                className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition hover:bg-muted"
              >
                {linkCopied ? "Link copied ✓" : "Invite link"}
              </button>
              <button
                type="button"
                onClick={() => setInviting(true)}
                className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
              >
                + Add user
              </button>
            </>
          )}
        </div>
      </header>

      {inviting && <InviteUserModal onClose={() => setInviting(false)} />}
      {auditOpen && <AuditLogModal onClose={() => setAuditOpen(false)} />}
      {editing && selected && <EditAccessModal user={selected} onClose={() => setEditing(false)} />}
      {enrolling && <Enroll2FAModal onClose={() => setEnrolling(false)} />}

      <div className="flex-1 space-y-4 overflow-auto p-4">
        {/* Stat cards */}
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <StatCard
            label="Active now"
            value={String(stats.data?.active_now ?? "—")}
            meta="on floor · scanning"
          />
          <StatCard
            label="Pending invites"
            value={String(stats.data?.pending_invites ?? "—")}
            meta="awaiting verification"
          />
          <StatCard
            label="2FA coverage"
            value={stats.data ? `${stats.data.twofa_coverage}%` : "—"}
            meta="of members"
          />
        </div>

        <div className="grid grid-cols-1 items-start gap-4 xl:grid-cols-[1fr_320px]">
          <div className="space-y-4">
            {/* Role tabs */}
            <div className="inline-flex flex-wrap gap-1 rounded-lg border border-border bg-surface p-1">
              {ROLE_TABS.map((t) => {
                const count =
                  t.key === "all" ? users.length : users.filter((u) => u.role === t.key).length;
                return (
                  <button
                    key={t.key}
                    type="button"
                    onClick={() => setFilter(t.key)}
                    className={`rounded-md px-2.5 py-1 text-xs font-medium transition ${
                      filter === t.key
                        ? "bg-primary text-white"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    {t.label}
                    {t.key === "all" ? ` · ${count}` : ""}
                  </button>
                );
              })}
            </div>

            {/* Roster table */}
            <div className="overflow-hidden rounded-lg border border-border bg-surface">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left text-xs uppercase tracking-wide text-muted-foreground">
                    <th className="px-3 py-2 font-medium">Member</th>
                    <th className="px-3 py-2 font-medium">Role</th>
                    <th className="px-3 py-2 font-medium">Zones</th>
                    <th className="px-3 py-2 font-medium">Shift</th>
                    <th className="px-3 py-2 font-medium">Status</th>
                    <th className="px-3 py-2 font-medium">Last seen</th>
                  </tr>
                </thead>
                <tbody>
                  {usersQuery.isLoading ? (
                    <tr>
                      <td colSpan={6} className="px-3 py-4 text-muted-foreground">
                        Loading users…
                      </td>
                    </tr>
                  ) : (
                    filtered.map((u) => (
                      <tr
                        key={u.id}
                        data-testid="user-row"
                        onClick={() => setSelectedId(u.id)}
                        className={`cursor-pointer border-t border-border transition hover:bg-muted/40 ${
                          selected?.id === u.id ? "bg-muted/50" : ""
                        }`}
                      >
                        <td className="px-3 py-2">
                          <div className="flex items-center gap-2.5">
                            <Avatar name={displayLabel(u)} />
                            <div className="min-w-0">
                              <div className="truncate font-medium text-foreground">
                                {displayLabel(u)}
                              </div>
                              <div className="truncate text-[11.5px] text-muted-foreground">
                                {u.title ? `${u.title} · ` : ""}
                                {u.email}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-3 py-2">
                          <RoleChip role={u.role} />
                        </td>
                        <td className="px-3 py-2 font-mono text-xs text-foreground">
                          {zonesLabel(u.zones)}
                        </td>
                        <td className="px-3 py-2 capitalize text-foreground">{u.shift}</td>
                        <td className="px-3 py-2">
                          <StatusChip status={u.status} />
                        </td>
                        <td className="px-3 py-2 font-mono text-xs text-muted-foreground">
                          {relativeTime(u.last_seen_at)}
                        </td>
                      </tr>
                    ))
                  )}
                  {!usersQuery.isLoading && filtered.length === 0 && (
                    <tr>
                      <td colSpan={6} className="px-3 py-4 text-center text-muted-foreground">
                        No members in this role.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Role permissions */}
            <div className="rounded-lg border border-border bg-surface p-4">
              <div className="mb-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Role permissions
              </div>
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="text-xs uppercase tracking-wide text-muted-foreground">
                    <th className="py-1 pr-3 font-medium">Capability</th>
                    {ROLE_COLUMNS.map((r) => (
                      <th key={r} className="px-2 py-1 text-center font-medium">
                        {r}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {PERMISSION_MATRIX.map((row) => (
                    <tr key={row.label} className="border-t border-border">
                      <td className="py-1.5 pr-3 text-foreground">{row.label}</td>
                      {row.allow.map((ok, i) => (
                        <td key={ROLE_COLUMNS[i]} className="px-2 py-1.5 text-center">
                          <span className={ok ? "text-primary" : "text-muted-foreground/40"}>
                            {ok ? "✓" : "—"}
                          </span>
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Side detail panel */}
          {selected && (
            <aside className="rounded-lg border border-border bg-surface xl:sticky xl:top-4">
              <div className="flex items-center gap-3 border-b border-border p-4">
                <Avatar name={displayLabel(selected)} size={44} />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold text-foreground">
                    {displayLabel(selected)}
                  </div>
                  <div className="truncate text-xs text-muted-foreground">
                    {selected.title || selected.role}
                  </div>
                </div>
                <StatusChip status={selected.status} />
              </div>
              <div className="space-y-3 p-4 text-sm">
                <Row k="Email" v={<span className="font-mono text-xs">{selected.email}</span>} />
                <Row k="Role" v={<RoleChip role={selected.role} />} />
                <Row k="Zone access" v={zonesLabel(selected.zones)} />
                <Row k="Shift" v={<span className="capitalize">{selected.shift}</span>} />
                <Row
                  k="Last seen"
                  v={
                    <span className="font-mono text-xs">{relativeTime(selected.last_seen_at)}</span>
                  }
                />
                <Row
                  k="2FA"
                  v={
                    selected.mfa_enabled ? (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-success/15 px-2 py-0.5 text-xs text-success">
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        Enabled
                      </span>
                    ) : me.data?.user_id === selected.id ? (
                      <button
                        type="button"
                        onClick={() => setEnrolling(true)}
                        className="text-xs font-medium text-primary hover:underline"
                      >
                        Set up authenticator
                      </button>
                    ) : (
                      <span className="text-xs text-muted-foreground">Not enabled</span>
                    )
                  }
                />

                <RecentActivity userId={selected.id} />

                {showInvite && selected.status === "pending" && (
                  <div className="mt-2 flex items-center justify-between gap-2 border-t border-border pt-3">
                    <span className="text-xs text-muted-foreground">Invitation pending</span>
                    <button
                      type="button"
                      disabled={resendInvite.isPending || !selected.idp_user_id}
                      onClick={() => resendInvite.mutate(selected.idp_user_id)}
                      className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition hover:bg-muted disabled:opacity-50"
                    >
                      {resendInvite.isSuccess ? "Invite resent ✓" : "Resend invite"}
                    </button>
                  </div>
                )}

                {showInvite && (
                  <div className="mt-2 flex items-center justify-between gap-2 border-t border-border pt-3">
                    <button
                      type="button"
                      disabled={resetPassword.isPending || !selected.idp_user_id}
                      onClick={() => resetPassword.mutate(selected.idp_user_id)}
                      className="text-sm font-medium text-foreground hover:underline disabled:opacity-50"
                    >
                      {resetPassword.isSuccess ? "Reset sent ✓" : "Reset password"}
                    </button>
                    <button
                      type="button"
                      onClick={() => setEditing(true)}
                      className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white hover:opacity-90"
                    >
                      Edit access
                    </button>
                  </div>
                )}
              </div>
            </aside>
          )}
        </div>
      </div>
    </div>
  );
}

function RecentActivity({ userId }: { userId: string }) {
  const audit = useAudit(userId, 6);
  const rows = audit.data ?? [];
  return (
    <div className="pt-2">
      <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
        Recent activity
      </div>
      {rows.length === 0 ? (
        <div className="text-xs text-muted-foreground">No recent activity.</div>
      ) : (
        <ul className="space-y-1.5">
          {rows.map((e, i) => (
            <li key={i} className="flex items-center justify-between gap-2 text-xs">
              <span className="flex items-center gap-2 truncate text-foreground">
                <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-primary" />
                {auditLabel(e.action)}
              </span>
              <span className="shrink-0 font-mono text-muted-foreground">{relativeTime(e.ts)}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function Row({ k, v }: { k: string; v: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-xs text-muted-foreground">{k}</span>
      <span className="text-right text-foreground">{v}</span>
    </div>
  );
}
