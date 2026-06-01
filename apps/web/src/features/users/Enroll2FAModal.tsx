import { useEffect, useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { useStartTotp, useVerifyTotp } from "./useUsers";

interface Enroll2FAModalProps {
  onClose: () => void;
  onEnrolled?: () => void;
}

/** Self-service authenticator (TOTP) enrollment: scan QR, enter first code. */
export function Enroll2FAModal({ onClose, onEnrolled }: Enroll2FAModalProps) {
  const start = useStartTotp();
  const verify = useVerifyTotp();
  const [code, setCode] = useState("");

  // Begin enrollment once when the modal opens.
  const { mutate: begin } = start;
  useEffect(() => {
    begin();
  }, [begin]);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    verify.mutate(code.trim(), { onSuccess: () => onEnrolled?.() });
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Set up authenticator"
    >
      <div className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-5 shadow-lg">
        <h2 className="text-base font-semibold text-foreground">Set up authenticator</h2>

        {verify.isSuccess ? (
          <>
            <p role="status" className="text-sm text-muted-foreground">
              Authenticator enrolled. You will be prompted for a code at next sign-in.
            </p>
            <div className="flex justify-end">
              <button
                type="button"
                onClick={onClose}
                className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white"
              >
                Done
              </button>
            </div>
          </>
        ) : (
          <>
            <p className="text-xs text-muted-foreground">
              Scan with Google Authenticator, 1Password, or Authy. Then enter the 6-digit code to
              confirm.
            </p>

            <div className="flex justify-center rounded-lg border border-border bg-white p-4">
              {start.isPending && <span className="text-sm text-muted-foreground">Loading…</span>}
              {start.isError && (
                <span className="text-sm text-destructive">Could not start enrollment.</span>
              )}
              {start.data && <QRCodeSVG value={start.data.uri} size={176} />}
            </div>

            {start.data && (
              <p className="text-center text-[11px] text-muted-foreground">
                Can&apos;t scan? Enter this key manually:{" "}
                <span className="font-mono text-foreground">{start.data.secret}</span>
              </p>
            )}

            <form onSubmit={submit} className="space-y-3">
              <label className="block text-sm">
                <span className="mb-1 block text-muted-foreground">6-digit code</span>
                <input
                  inputMode="numeric"
                  pattern="[0-9]*"
                  maxLength={6}
                  required
                  value={code}
                  onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
                  placeholder="123456"
                  className="w-full rounded border border-border bg-background px-3 py-2 text-center font-mono text-lg tracking-widest text-foreground"
                />
              </label>

              {verify.isError && (
                <p role="alert" className="text-xs text-destructive">
                  Invalid code. Check your authenticator and try again.
                </p>
              )}

              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={onClose}
                  className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={verify.isPending || code.length !== 6 || !start.data}
                  className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
                >
                  {verify.isPending ? "Verifying…" : "Verify"}
                </button>
              </div>
            </form>
          </>
        )}
      </div>
    </div>
  );
}
