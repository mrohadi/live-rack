import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { SignalCards } from "../SignalCards";
import { addRecommendation } from "../useRecommendations";
import { isRecommendation, type Recommendation } from "../../../lib/ws";

function rec(id: string, title = "Move rain gear"): Recommendation {
  return {
    id,
    org_id: "o",
    store_id: "s",
    kind: "weather",
    severity: "action",
    title,
    rationale: "Rain in NYC",
    suggested_task: "Stock umbrellas",
    created_at: "2026-06-01T00:00:00Z",
  };
}

describe("isRecommendation", () => {
  it("accepts recommendation frames and rejects scan frames", () => {
    expect(isRecommendation(rec("1"))).toBe(true);
    expect(isRecommendation({ scanner_id: "Z", action: "pick" })).toBe(false);
    expect(isRecommendation(null)).toBe(false);
  });
});

describe("addRecommendation", () => {
  it("prepends and dedupes by id", () => {
    const list = addRecommendation([rec("1")], rec("2"));
    expect(list.map((r) => r.id)).toEqual(["2", "1"]);
    const deduped = addRecommendation(list, rec("1", "updated"));
    expect(deduped.map((r) => r.id)).toEqual(["1", "2"]);
    expect(deduped[0].title).toBe("updated");
  });
});

describe("SignalCards", () => {
  it("renders cards and fires onApply", () => {
    const onApply = vi.fn();
    render(<SignalCards recommendations={[rec("1")]} onApply={onApply} />);
    expect(screen.getByText("Move rain gear")).toBeInTheDocument();
    expect(screen.getByText("→ Stock umbrellas")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Apply" }));
    expect(onApply).toHaveBeenCalledOnce();
  });

  it("shows applying state and disables the button", () => {
    render(<SignalCards recommendations={[rec("1")]} onApply={vi.fn()} applyingId="1" />);
    const btn = screen.getByRole("button", { name: "Applying…" });
    expect(btn).toBeDisabled();
  });

  it("shows empty state", () => {
    render(<SignalCards recommendations={[]} onApply={vi.fn()} />);
    expect(screen.getByText(/no active signals/i)).toBeInTheDocument();
  });
});
