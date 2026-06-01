import { useState } from "react";
import {
  ASSIGNABLE_ROLES,
  useInviteUser,
  type AssignableRole,
  type InviteResult,
} from "./useUsers";

interface InviteUserModalProps {
  onClose: () => void;
  onInvited?: (r: InviteResult) => void;
}

/** Modal form for inviting a teammate. Admin + 2FA gating happens upstream. */
export function InviteUserModal({ onClose, onInvited }: InviteUserModalProps) {
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [role, setRole] = useState<AssignableRole>("staff");
  const invite = useInviteUser();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    invite.mutate(
      { email: email.trim(), display_name: displayName.trim(), role },
      { onSuccess: (r) => onInvited?.(r) },
    );
  };

  if (invite.isSuccess) {
    return (
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
        role="dialog"
        aria-modal="true"
        aria-label="Invite user"
      >
        <div className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg">
          <div className="flex items-center gap-2">
            <span className="flex h-7 w-7 items-center justify-center rounded-full bg-emerald-100 text-sm font-bold text-emerald-700">
              ✓
            </span>
            <h2 className="text-base font-semibold text-foreground">Invitation sent</h2>
          </div>
          <p role="status" className="text-sm text-muted-foreground">
            Verification link emailed to <strong>{invite.data.email}</strong> as{" "}
            <strong>{invite.data.role}</strong>. They confirm the address, set a password, and enrol
            2FA before first sign-in.
          </p>
          <div className="flex justify-end pt-1">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white"
            >
              Done
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Invite user"
    >
      <form
        onSubmit={submit}
        className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <h2 className="text-base font-semibold text-foreground">Invite a teammate</h2>
        <p className="text-xs text-muted-foreground">
          We email a verification link. They confirm the address, set a password, and enrol 2FA
          before first sign-in.
        </p>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Email</span>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="teammate@company.com"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Display name</span>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="Ada Lovelace"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Role</span>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as AssignableRole)}
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm capitalize text-foreground"
          >
            {ASSIGNABLE_ROLES.map((r) => (
              <option key={r} value={r}>
                {r}
              </option>
            ))}
          </select>
        </label>

        {invite.isError && (
          <p role="alert" className="text-xs text-destructive">
            Invite failed: {invite.error.message}. Check the email and try again.
          </p>
        )}

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={invite.isPending}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            {invite.isPending ? "Sending…" : "Send invite"}
          </button>
        </div>
      </form>
    </div>
  );
}
