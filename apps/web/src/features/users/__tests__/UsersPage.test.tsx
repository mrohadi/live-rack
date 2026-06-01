import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const get = vi.fn();
const post = vi.fn().mockResolvedValue(undefined);
vi.mock("../../../lib/api", () => ({ useApi: () => ({ get, post }) }));
vi.mock("react-oidc-context", () => ({
  useAuth: () => ({ user: { profile: { amr: ["pwd", "otp"] } } }),
}));

import { UsersPage } from "../UsersPage";

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <UsersPage />
    </QueryClientProvider>,
  );
}

describe("UsersPage", () => {
  beforeEach(() => {
    get.mockReset();
    get.mockImplementation((path: string) => {
      if (path === "/api/v1/users") {
        return Promise.resolve([
          {
            id: "1",
            email: "ann@x.io",
            display_name: "Ann Lee",
            avatar_url: "",
            role: "admin",
            title: "Ops Manager",
            shift: "day",
            status: "active",
            mfa_enabled: true,
            last_seen_at: new Date().toISOString(),
            zones: [],
          },
        ]);
      }
      if (path === "/api/v1/users/stats") {
        return Promise.resolve({
          members: 1,
          roles: 5,
          active_now: 1,
          pending_invites: 0,
          twofa_coverage: 100,
        });
      }
      return Promise.resolve({
        user_id: "u",
        role: "admin",
        mfa_verified: true,
        permissions: ["edit_users"],
        store_scoped: false,
        zone_scoped: false,
      });
    });
  });
  afterEach(() => vi.clearAllMocks());

  it("renders roster and the permission matrix", async () => {
    renderPage();
    // Ann Lee appears in both the roster row and the side detail panel.
    await waitFor(() => expect(screen.getAllByText("Ann Lee").length).toBeGreaterThan(0));
    expect(screen.getByText("Edit users")).toBeInTheDocument();
    expect(screen.getByText("Manage integrations")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText(/you: admin · 2FA on/)).toBeInTheDocument());
  });
});
