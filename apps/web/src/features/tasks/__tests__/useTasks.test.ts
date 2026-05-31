import { describe, expect, it } from "vitest";
import type { Task, TaskStatus } from "../types";
import { groupByStatus, moveTask } from "../useTasks";

function task(over: Partial<Task>): Task {
  return {
    id: "t1",
    store_id: "s1",
    title: "Restock",
    status: "todo",
    priority: "med",
    updated_at: "t0",
    ...over,
  };
}

describe("groupByStatus", () => {
  it("buckets tasks into all four columns", () => {
    const g = groupByStatus([
      task({ id: "a", status: "todo" }),
      task({ id: "b", status: "done" }),
      task({ id: "c", status: "todo" }),
    ]);
    expect(g.todo.map((t) => t.id)).toEqual(["a", "c"]);
    expect(g.done.map((t) => t.id)).toEqual(["b"]);
    expect(g.in_progress).toEqual([]);
    expect(g.review).toEqual([]);
  });
});

describe("moveTask", () => {
  it("updates the status of the matching task", () => {
    const next = moveTask([task({ id: "a", status: "todo" })], "a", "done");
    expect(next[0].status).toBe<TaskStatus>("done");
  });

  it("is a no-op when status is unchanged", () => {
    const rows = [task({ id: "a", status: "todo" })];
    const next = moveTask(rows, "a", "todo");
    expect(next[0]).toBe(rows[0]);
  });

  it("leaves non-matching tasks untouched", () => {
    const next = moveTask([task({ id: "a", status: "todo" })], "zzz", "done");
    expect(next[0].status).toBe<TaskStatus>("todo");
  });
});
