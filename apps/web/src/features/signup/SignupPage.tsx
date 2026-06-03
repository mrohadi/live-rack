import { useState } from "react";
import { Link } from "react-router-dom";
import { AuthLayout } from "../../components/auth/AuthLayout";
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
      <AuthLayout title="Check your email" subtitle="Your workspace is ready.">
        <p className="text-sm text-muted-foreground">
          We created your workspace. Verify your email and set a password to finish signing up.
        </p>
        <Link
          to="/"
          className="mt-6 block w-full rounded-md bg-primary px-3 py-2.5 text-center text-sm font-medium text-white transition hover:opacity-90"
        >
          Go to sign in
        </Link>
      </AuthLayout>
    );
  }

  const field =
    "w-full rounded-md border border-border bg-surface px-3 py-2.5 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20";

  return (
    <AuthLayout title="Create your workspace" subtitle="Start a new live-rack organization.">
      <form onSubmit={submit} className="space-y-4">
        <label className="block text-sm">
          <span className="mb-1.5 block font-medium text-foreground">Company</span>
          <input
            type="text"
            required
            value={company}
            onChange={(e) => setCompany(e.target.value)}
            placeholder="Acme Co"
            className={field}
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1.5 block font-medium text-foreground">Work email</span>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            className={field}
          />
        </label>

        <label className="block text-sm">
          <span className="mb-1.5 block font-medium text-foreground">Your name</span>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="Ada Lovelace"
            className={field}
          />
        </label>

        {signup.isError && (
          <p role="alert" className="text-xs text-destructive">
            {signup.error instanceof Error && signup.error.message === "email_taken"
              ? "That email is already registered. Try signing in instead."
              : "Signup failed. Please try again."}
          </p>
        )}

        <button
          type="submit"
          disabled={signup.isPending}
          className="w-full rounded-md bg-primary px-3 py-2.5 text-sm font-medium text-white transition hover:opacity-90 disabled:opacity-50"
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
    </AuthLayout>
  );
}
