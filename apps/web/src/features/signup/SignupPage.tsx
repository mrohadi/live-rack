import { useState } from "react";
import { Link } from "react-router-dom";
import { useSignup } from "./useSignup";

/** Public self-service registration. Creates a tenant org + admin, then prompts
 *  the user to verify their email and sign in. */
export function SignupPage() {
  const [company, setCompany] = useState("");
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const signup = useSignup();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    signup.mutate({
      company: company.trim(),
      email: email.trim(),
      display_name: displayName.trim(),
    });
  };

  if (signup.isSuccess) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <div className="w-full max-w-md space-y-3 rounded-lg border border-border bg-surface p-6 text-center">
          <h1 className="text-lg font-semibold text-foreground">Check your email</h1>
          <p className="text-sm text-muted-foreground">
            We created your workspace. Verify your email and set a password to finish signing up.
          </p>
          <Link to="/" className="inline-block text-sm font-medium text-primary">
            Go to sign in →
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <form
        onSubmit={submit}
        className="w-full max-w-md space-y-4 rounded-lg border border-border bg-surface p-6"
      >
        <div>
          <h1 className="text-lg font-semibold text-foreground">Create your workspace</h1>
          <p className="text-xs text-muted-foreground">Start a new live-rack organization.</p>
        </div>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Company</span>
          <input
            type="text"
            required
            value={company}
            onChange={(e) => setCompany(e.target.value)}
            placeholder="Acme Co"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Work email</span>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1 block text-muted-foreground">Your name</span>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="Ada Lovelace"
            className="w-full rounded border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </label>

        {signup.isError && (
          <p role="alert" className="text-xs text-destructive">
            Signup failed. That email or company may already exist.
          </p>
        )}

        <button
          type="submit"
          disabled={signup.isPending}
          className="w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
        >
          {signup.isPending ? "Creating…" : "Create workspace"}
        </button>

        <p className="text-center text-xs text-muted-foreground">
          Already have an account?{" "}
          <Link to="/" className="font-medium text-primary">
            Sign in
          </Link>
        </p>
      </form>
    </div>
  );
}
