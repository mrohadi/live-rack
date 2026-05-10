import { describe, expect, test } from "vitest";
import {
  REQUIRED_COLOR_TOKENS,
  REQUIRED_STATIC_TOKENS,
  lightTokens,
  darkTokens,
  staticTokens,
} from "./tokens";

const HSL_RE = /^\d{1,3} \d{1,3}% \d{1,3}%$/;

describe("design tokens", () => {
  test("all required color tokens defined in light mode", () => {
    for (const token of REQUIRED_COLOR_TOKENS) {
      expect(lightTokens[token], `missing light token: ${token}`).toBeDefined();
      expect(
        lightTokens[token],
        `invalid HSL for light token: ${token}`,
      ).toMatch(HSL_RE);
    }
  });

  test("all required color tokens defined in dark mode", () => {
    for (const token of REQUIRED_COLOR_TOKENS) {
      expect(darkTokens[token], `missing dark token: ${token}`).toBeDefined();
      expect(
        darkTokens[token],
        `invalid HSL for dark token: ${token}`,
      ).toMatch(HSL_RE);
    }
  });

  test("light and dark token sets have identical keys", () => {
    expect(Object.keys(lightTokens).sort()).toEqual(
      Object.keys(darkTokens).sort(),
    );
  });

  test("static tokens defined (font, radius)", () => {
    for (const token of REQUIRED_STATIC_TOKENS) {
      expect(
        staticTokens[token],
        `missing static token: ${token}`,
      ).toBeDefined();
      expect(staticTokens[token].length).toBeGreaterThan(0);
    }
  });
});
