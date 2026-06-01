import { describe, expect, it } from "vitest";
import {
  ASSIGNABLE_ROLES,
  PERMISSION_MATRIX,
  ROLE_COLUMNS,
  canInvite,
  initials,
} from "../useUsers";

describe("initials", () => {
  it("takes up to two leading initials", () => {
    expect(initials("Avery Chen")).toBe("AC");
    expect(initials("madonna")).toBe("M");
    expect(initials("  Jon  Ola  Vik ")).toBe("JO");
    expect(initials("")).toBe("");
  });
});

describe("PERMISSION_MATRIX", () => {
  it("has one allow flag per role column", () => {
    for (const row of PERMISSION_MATRIX) {
      expect(row.allow).toHaveLength(ROLE_COLUMNS.length);
    }
  });

  it("encodes the admin-only and read-only rules from the design", () => {
    const editUsers = PERMISSION_MATRIX.find((r) => r.label === "Edit users")!;
    expect(editUsers.allow).toEqual([true, false, false, false]);
    const exportReports = PERMISSION_MATRIX.find((r) => r.label === "Export reports")!;
    expect(exportReports.allow[3]).toBe(true); // read-only can export
  });
});

describe("canInvite", () => {
  const base = {
    user_id: "u",
    permissions: [],
    store_scoped: false,
    zone_scoped: false,
  };
  it("allows only admins with a verified second factor", () => {
    expect(canInvite({ ...base, role: "admin", mfa_verified: true })).toBe(true);
    expect(canInvite({ ...base, role: "admin", mfa_verified: false })).toBe(false);
    expect(canInvite({ ...base, role: "manager", mfa_verified: true })).toBe(false);
    expect(canInvite(undefined)).toBe(false);
  });
});

describe("ASSIGNABLE_ROLES", () => {
  it("excludes the service role", () => {
    expect(ASSIGNABLE_ROLES).not.toContain("service");
    expect(ASSIGNABLE_ROLES).toContain("admin");
  });
});
