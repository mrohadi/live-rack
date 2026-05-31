import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const get = vi.fn();
vi.mock("../../../lib/api", () => ({ useApi: () => ({ get }) }));

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
          { id: "1", email: "ann@x.io", display_name: "Ann Lee", avatar_url: "", role: "admin" },
        ]);
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
    await waitFor(() => expect(screen.getByText("Ann Lee")).toBeInTheDocument());
    expect(screen.getByText("Edit users")).toBeInTheDocument();
    expect(screen.getByText("Manage integrations")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText(/you: admin · 2FA on/)).toBeInTheDocument());
  });
});
