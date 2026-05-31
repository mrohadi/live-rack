// AuthProvider processes the OIDC redirect automatically; this just shows a spinner.
export function CallbackPage() {
  return (
    <div className="flex h-screen items-center justify-center">
      <span className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
    </div>
  );
}
