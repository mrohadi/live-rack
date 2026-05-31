import { describe, expect, it } from "vitest";
import { PERMISSION_MATRIX, ROLE_COLUMNS, initials } from "../useUsers";

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
