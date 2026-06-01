import { useState } from "react";
import { useAuth } from "react-oidc-context";
import { InviteUserModal } from "./InviteUserModal";
import {
  PERMISSION_MATRIX,
  ROLE_COLUMNS,
  canInvite,
  hasMfa,
  initials,
  useCapabilities,
  useUsers,
  type OrgUser,
} from "./useUsers";

function RoleChip({ role }: { role: string }) {
  return (
    <span className="rounded bg-border px-1.5 py-0.5 text-[11px] font-medium capitalize text-foreground">
      {role || "—"}
    </span>
  );
}

function UserRow({ u }: { u: OrgUser }) {
  return (
    <tr className="border-t border-border">
      <td className="px-3 py-2">
        <div className="flex items-center gap-2">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-border text-[11px] font-semibold text-foreground">
            {initials(u.display_name)}
          </span>
          <div className="min-w-0">
            <div className="truncate text-sm font-medium text-foreground">{u.display_name}</div>
            <div className="truncate text-xs text-muted-foreground">{u.email}</div>
          </div>
        </div>
      </td>
      <td className="px-3 py-2">
        <RoleChip role={u.role} />
      </td>
    </tr>
  );
}

export function UsersPage() {
  const usersQuery = useUsers();
  const me = useCapabilities();
  const auth = useAuth();
  const users = usersQuery.data ?? [];
  const [inviting, setInviting] = useState(false);
  const showInvite = canInvite(me.data);
  const mfaOn = hasMfa(auth.user?.profile);

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <h1 className="text-lg font-semibold text-foreground">Users &amp; Access</h1>
          <p className="text-xs text-muted-foreground">
            {users.length} members
            {me.data ? ` · you: ${me.data.role}${mfaOn ? " · 2FA on" : ""}` : ""}
          </p>
        </div>
        {showInvite && (
          <button
            type="button"
            onClick={() => setInviting(true)}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground"
          >
            Invite member
          </button>
        )}
      </header>

      {inviting && <InviteUserModal onClose={() => setInviting(false)} />}

      <div className="flex-1 space-y-4 overflow-auto p-4">
        <div className="overflow-hidden rounded-lg border border-border bg-surface">
          <table className="w-full text-left">
            <thead>
              <tr className="text-xs uppercase tracking-wide text-muted-foreground">
                <th className="px-3 py-2 font-medium">Member</th>
                <th className="px-3 py-2 font-medium">Role</th>
              </tr>
            </thead>
            <tbody>
              {usersQuery.isLoading ? (
                <tr>
                  <td className="px-3 py-3 text-sm text-muted-foreground" colSpan={2}>
                    Loading users…
                  </td>
                </tr>
              ) : (
                users.map((u) => <UserRow key={u.id} u={u} />)
              )}
            </tbody>
          </table>
        </div>

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
    </div>
  );
}
