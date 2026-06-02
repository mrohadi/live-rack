import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ToastProvider } from "../toast";
import { useToast } from "../toast-context";
import { ConfirmProvider } from "../confirm";
import { useConfirm } from "../confirm-context";

function ToastButton() {
  const toast = useToast();
  return (
    <button type="button" onClick={() => toast.success("Saved")}>
      fire
    </button>
  );
}

function ConfirmButton({ onResult }: { onResult: (ok: boolean) => void }) {
  const confirm = useConfirm();
  // Fire-and-forget so the click handler does not block on the promise.
  const ask = () => {
    void confirm({ title: "Delete zone", confirmLabel: "Delete" }).then(onResult);
  };
  return (
    <button type="button" onClick={ask}>
      ask
    </button>
  );
}

describe("ToastProvider", () => {
  it("shows a toast on success", () => {
    render(
      <ToastProvider>
        <ToastButton />
      </ToastProvider>,
    );
    fireEvent.click(screen.getByText("fire"));
    expect(screen.getByText("Saved")).toBeInTheDocument();
  });

  it("auto-dismisses after the timeout", () => {
    vi.useFakeTimers();
    render(
      <ToastProvider>
        <ToastButton />
      </ToastProvider>,
    );
    fireEvent.click(screen.getByText("fire"));
    expect(screen.getByText("Saved")).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(4000);
    });
    expect(screen.queryByText("Saved")).not.toBeInTheDocument();
  });
});

afterEach(() => {
  vi.useRealTimers();
});

describe("ConfirmProvider", () => {
  it("resolves true on confirm", async () => {
    const onResult = vi.fn();
    render(
      <ConfirmProvider>
        <ConfirmButton onResult={onResult} />
      </ConfirmProvider>,
    );
    fireEvent.click(screen.getByText("ask"));
    fireEvent.click(screen.getByRole("button", { name: "Delete" }));
    await waitFor(() => expect(onResult).toHaveBeenCalledWith(true));
  });

  it("resolves false on cancel", async () => {
    const onResult = vi.fn();
    render(
      <ConfirmProvider>
        <ConfirmButton onResult={onResult} />
      </ConfirmProvider>,
    );
    fireEvent.click(screen.getByText("ask"));
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    await waitFor(() => expect(onResult).toHaveBeenCalledWith(false));
  });
});
