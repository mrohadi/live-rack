import { describe, expect, it } from "vitest";
import { resetErrorMessage } from "../usePasswordReset";

describe("resetErrorMessage", () => {
  it("maps known errors", () => {
    expect(resetErrorMessage(new Error("400: invalid or expired code"))).toBe(
      "This link is invalid or has expired.",
    );
    expect(resetErrorMessage(new Error("400: password too short"))).toBe(
      "Please check the details and try again.",
    );
    expect(resetErrorMessage("boom")).toBe("Something went wrong. Try again.");
  });
});
