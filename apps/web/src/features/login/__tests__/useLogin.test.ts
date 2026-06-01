import { describe, expect, it } from "vitest";
import { loginErrorMessage } from "../useLogin";

describe("loginErrorMessage", () => {
  it("maps 401 to a credentials message", () => {
    expect(loginErrorMessage(new Error("401: invalid credentials"))).toMatch(/incorrect email/i);
  });

  it("maps 400 to a details message", () => {
    expect(loginErrorMessage(new Error("400: bad"))).toMatch(/check the details/i);
  });

  it("falls back for unknown errors", () => {
    expect(loginErrorMessage(new Error("500: boom"))).toMatch(/something went wrong/i);
    expect(loginErrorMessage("weird")).toMatch(/something went wrong/i);
  });
});
