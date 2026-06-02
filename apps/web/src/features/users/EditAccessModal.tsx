import { useState } from "react";
import { useToast } from "../../components/feedback/toast-context";
import { ASSIGNABLE_ROLES, useSetRole, type AssignableRole, type OrgUser } from "./useUsers";

interface EditAccessModalProps {
  user: OrgUser;
  onClose: () => void;
}

/** Modal to change a member's role. Re-grants in Zitadel via the idp id. */
export function EditAccessModal({ user, onClose }: EditAccessModalProps) {
  const [role, setRole] = useState<AssignableRole>(
    (ASSIGNABLE_ROLES as readonly string[]).includes(user.role)
      ? (user.role as AssignableRole)
      : "staff",
  );
  const setRoleMut = useSetRole();
  const toast = useToast();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setRoleMut.mutate(
      { id: user.id, idpUserId: user.idp_user_id, role },
      {
        onSuccess: () => {
          toast.success("Role updated");
          onClose();
        },
        onError: () => toast.error("Failed to update role"),
      },
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Edit access"
    >
      <form
        onSubmit={submit}
        className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <h2 className="text-base font-semibold text-foreground">
          Edit access · {user.display_name}
        </h2>

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

        {setRoleMut.isError && (
          <p role="alert" className="text-xs text-destructive">
            Could not update role. Try again.
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
            disabled={setRoleMut.isPending}
            className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            {setRoleMut.isPending ? "Saving…" : "Save access"}
          </button>
        </div>
      </form>
    </div>
  );
}
