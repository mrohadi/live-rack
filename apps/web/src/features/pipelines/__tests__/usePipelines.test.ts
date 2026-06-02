import { describe, expect, it } from "vitest";
import type { Card } from "../types";
import { ageingCount, cardsByStage, formatAge, moveCard } from "../usePipelines";

function card(over: Partial<Card>): Card {
  return {
    id: "c1",
    stage_position: 0,
    title: "Espresso Machine",
    priority: "medium",
    entered_stage_at: "2026-05-31T00:00:00Z",
    age_seconds: 0,
    ageing: false,
    ...over,
  };
}

describe("cardsByStage", () => {
  it("buckets cards by stage position, seeding every stage", () => {
    const g = cardsByStage(
      [0, 1, 2],
      [
        card({ id: "a", stage_position: 0 }),
        card({ id: "b", stage_position: 2 }),
        card({ id: "c", stage_position: 0 }),
      ],
    );
    expect(g[0].map((c) => c.id)).toEqual(["a", "c"]);
    expect(g[1]).toEqual([]);
    expect(g[2].map((c) => c.id)).toEqual(["b"]);
  });
});

describe("moveCard", () => {
  it("moves a card and resets its ageing state", () => {
    const next = moveCard(
      [card({ id: "a", stage_position: 0, age_seconds: 9000, ageing: true })],
      "a",
      1,
    );
    expect(next[0].stage_position).toBe(1);
    expect(next[0].age_seconds).toBe(0);
    expect(next[0].ageing).toBe(false);
  });

  it("is a no-op when the stage is unchanged", () => {
    const rows = [card({ id: "a", stage_position: 0 })];
    expect(moveCard(rows, "a", 0)[0]).toBe(rows[0]);
  });
});

describe("ageingCount", () => {
  it("counts cards breaching SLA", () => {
    expect(
      ageingCount([card({ ageing: true }), card({ ageing: false }), card({ ageing: true })]),
    ).toBe(2);
  });
});

describe("formatAge", () => {
  it("renders compact dwell times", () => {
    expect(formatAge(30)).toBe("30s");
    expect(formatAge(120)).toBe("2m");
    expect(formatAge(7200)).toBe("2h");
    expect(formatAge(172800)).toBe("2d");
  });
});
