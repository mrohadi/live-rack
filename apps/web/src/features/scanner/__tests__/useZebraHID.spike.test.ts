import { describe, expect, it } from "vitest";
import { decodeHIDReport } from "../useZebraHID";

describe("decodeHIDReport", () => {
  it("decodes a lowercase keycode", () => {
    expect(decodeHIDReport(0x00, 0x04)).toBe("a");
  });

  it("decodes uppercase with shift modifier", () => {
    expect(decodeHIDReport(0x02, 0x04)).toBe("A");
  });

  it("returns \\r for Enter keycode 0x28", () => {
    expect(decodeHIDReport(0x00, 0x28)).toBe("\r");
  });

  it("returns null for unknown keycode", () => {
    expect(decodeHIDReport(0x00, 0xff)).toBeNull();
  });
});
