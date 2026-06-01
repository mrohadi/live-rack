import { Link } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";

interface WelcomeProps {
  onSignIn: () => void;
}

/** Branded sign-in landing shown to signed-out users. "Sign in" hands off to the
 *  Zitadel hosted login; "Create workspace" routes to self-service signup. */
export function Welcome({ onSignIn }: WelcomeProps) {
  return (
    <AuthLayout title="Welcome back" subtitle="Sign in to your live-rack workspace.">
      <div className="space-y-3">
        <button
          type="button"
          onClick={onSignIn}
          className="w-full rounded-md bg-primary px-3 py-2.5 text-sm font-medium text-white transition hover:opacity-90"
        >
          Sign in
        </button>

        <div className="flex items-center gap-3 py-1 text-xs text-muted-foreground">
          <span className="h-px flex-1 bg-border" />
          new to live-rack?
          <span className="h-px flex-1 bg-border" />
        </div>

        <Link
          to="/signup"
          className="block w-full rounded-md border border-border px-3 py-2.5 text-center text-sm font-medium text-foreground transition hover:bg-muted"
        >
          Create a workspace
        </Link>
      </div>

      <p className="mt-8 text-center text-xs text-muted-foreground">
        Secured by Zitadel · SSO &amp; 2FA supported
      </p>
    </AuthLayout>
  );
}
