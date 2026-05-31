import { describe, expect, it } from "vitest";
import { MIN_QUERY_LENGTH, searchKeys, searchPath } from "../useSearch";

describe("searchPath", () => {
  it("encodes the query and default limit", () => {
    expect(searchPath("frozen bay")).toBe("/api/v1/search?q=frozen+bay&limit=20");
  });

  it("honours a custom limit", () => {
    expect(searchPath("wid", 5)).toBe("/api/v1/search?q=wid&limit=5");
  });
});

describe("searchKeys", () => {
  it("namespaces keys by query", () => {
    expect(searchKeys.query("wid")).toEqual(["search", "wid"]);
  });
});

describe("MIN_QUERY_LENGTH", () => {
  it("matches the API minimum", () => {
    expect(MIN_QUERY_LENGTH).toBe(2);
  });
});
