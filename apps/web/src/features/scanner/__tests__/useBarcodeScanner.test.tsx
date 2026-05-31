import { render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

const decodeFromVideoDevice = vi.fn();
const stopFn = vi.fn();

vi.mock("@zxing/browser", () => ({
  BrowserMultiFormatReader: vi.fn().mockImplementation(() => ({
    decodeFromVideoDevice,
  })),
}));

import { useBarcodeScanner } from "../useBarcodeScanner";

function Harness({ onScan }: { onScan: (sku: string) => void }) {
  const videoRef = useBarcodeScanner({ onScan, active: true });
  return <video ref={videoRef} />;
}

afterEach(() => vi.clearAllMocks());

describe("useBarcodeScanner", () => {
  it("invokes onScan with decoded text", () => {
    decodeFromVideoDevice.mockImplementation((_id, _el, cb) => {
      cb({ getText: () => "SKU-123" }, undefined);
      return Promise.resolve({ stop: stopFn });
    });
    const onScan = vi.fn();
    render(<Harness onScan={onScan} />);
    expect(onScan).toHaveBeenCalledWith("SKU-123");
  });

  it("ignores NotFound errors (no decode)", () => {
    decodeFromVideoDevice.mockImplementation((_id, _el, cb) => {
      cb(undefined, new Error("NotFoundException"));
      return Promise.resolve({ stop: stopFn });
    });
    const onScan = vi.fn();
    render(<Harness onScan={onScan} />);
    expect(onScan).not.toHaveBeenCalled();
  });
});
