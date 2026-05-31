import { describe, expect, it } from "vitest";
import { statusTone } from "../useIntegrations";

describe("statusTone", () => {
  it("maps healthy states to success tone", () => {
    expect(statusTone("connected")).toContain("success");
    expect(statusTone("processed")).toContain("success");
  });

  it("maps failure states to destructive tone", () => {
    expect(statusTone("error")).toContain("destructive");
    expect(statusTone("rejected")).toContain("destructive");
  });

  it("falls back to muted for unknown states", () => {
    expect(statusTone("received")).toContain("muted");
  });
});
