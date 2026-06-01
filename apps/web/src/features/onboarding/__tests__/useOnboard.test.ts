import { describe, expect, it } from "vitest";
import { onboardErrorMessage, passwordRules, passwordValid } from "../useOnboard";

describe("passwordRules", () => {
  it("flags each unmet requirement", () => {
    const rules = passwordRules("abc", "abc");
    const byLabel = Object.fromEntries(rules.map((r) => [r.label, r.ok]));
    expect(byLabel["At least 8 characters"]).toBe(false);
    expect(byLabel["An uppercase letter"]).toBe(false);
    expect(byLabel["A lowercase letter"]).toBe(true);
    expect(byLabel["A number"]).toBe(false);
    expect(byLabel["A symbol"]).toBe(false);
    expect(byLabel["Passwords match"]).toBe(true);
  });

  it("passes a strong, matching password", () => {
    expect(passwordValid("Sup3rSecret!", "Sup3rSecret!")).toBe(true);
  });

  it("fails when confirmation differs", () => {
    expect(passwordValid("Sup3rSecret!", "Sup3rSecret?")).toBe(false);
  });

  it("treats an empty password as non-matching", () => {
    const rules = passwordRules("", "");
    expect(rules.find((r) => r.label === "Passwords match")?.ok).toBe(false);
  });
});

describe("onboardErrorMessage", () => {
  it("maps known errors", () => {
    expect(onboardErrorMessage(new Error("400: invalid or expired code"))).toBe(
      "This link is invalid or has expired.",
    );
    expect(onboardErrorMessage(new Error("400: password rejected"))).toBe(
      "Please check the details and try again.",
    );
    expect(onboardErrorMessage("boom")).toBe("Something went wrong. Try again.");
  });
});
