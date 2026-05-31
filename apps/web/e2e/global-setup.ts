import type { FullConfig } from "@playwright/test";

// Zitadel needs no global bootstrap (Clerk's clerkSetup() is gone).
// Per-test auth is seeded via seedOidcSession() in ./auth.
export default async function globalSetup(_config: FullConfig) {}
