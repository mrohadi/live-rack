import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useNotificationCenter } from "../useNotificationCenter";
import type { TaskNotification } from "../ws";

const ME = "user-1";

vi.mock("../../features/users/useUsers", () => ({
  useCapabilities: () => ({ data: { user_id: ME } }),
}));

function notif(over: Partial<TaskNotification>): TaskNotification {
  return {
    org_id: "o1",
    store_id: "s1",
    task_id: "t1",
    assignee_id: ME,
    title: "Restock frozen",
    kind: "assigned",
    ts: new Date().toISOString(),
    ...over,
  };
}

describe("useNotificationCenter", () => {
  it("keeps notifications addressed to the current user", () => {
    const { result } = renderHook(() => useNotificationCenter());
    act(() => {
      expect(result.current.push(notif({}))).toBe(true);
    });
    expect(result.current.items).toHaveLength(1);
    expect(result.current.unread).toBe(1);
  });

  it("drops notifications for other assignees", () => {
    const { result } = renderHook(() => useNotificationCenter());
    act(() => {
      expect(result.current.push(notif({ assignee_id: "someone-else" }))).toBe(false);
    });
    expect(result.current.items).toHaveLength(0);
    expect(result.current.unread).toBe(0);
  });

  it("markAllRead zeroes the unread count but keeps history", () => {
    const { result } = renderHook(() => useNotificationCenter());
    act(() => void result.current.push(notif({ task_id: "a" })));
    act(() => void result.current.push(notif({ task_id: "b" })));
    expect(result.current.unread).toBe(2);
    act(() => result.current.markAllRead());
    expect(result.current.unread).toBe(0);
    expect(result.current.items).toHaveLength(2);
  });
});
