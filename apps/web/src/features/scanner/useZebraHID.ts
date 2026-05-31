import { useCallback, useEffect, useRef } from "react";

const ZEBRA_VENDOR_ID = 0x05e0;

const HID_KEYCODE_TO_CHAR: Record<number, string> = {
  0x04: "a",
  0x05: "b",
  0x06: "c",
  0x07: "d",
  0x08: "e",
  0x09: "f",
  0x0a: "g",
  0x0b: "h",
  0x0c: "i",
  0x0d: "j",
  0x0e: "k",
  0x0f: "l",
  0x10: "m",
  0x11: "n",
  0x12: "o",
  0x13: "p",
  0x14: "q",
  0x15: "r",
  0x16: "s",
  0x17: "t",
  0x18: "u",
  0x19: "v",
  0x1a: "w",
  0x1b: "x",
  0x1c: "y",
  0x1d: "z",
  0x1e: "1",
  0x1f: "2",
  0x20: "3",
  0x21: "4",
  0x22: "5",
  0x23: "6",
  0x24: "7",
  0x25: "8",
  0x26: "9",
  0x27: "0",
  0x28: "\r",
  0x2d: "-",
  0x2e: ".",
  0x2f: "/",
};

const SHIFT_KEYCODE_TO_CHAR: Record<number, string> = {
  0x1e: "!",
  0x1f: "@",
  0x20: "#",
  0x21: "$",
  0x22: "%",
  0x23: "^",
  0x24: "&",
  0x25: "*",
  0x26: "(",
  0x27: ")",
  0x04: "A",
  0x05: "B",
  0x06: "C",
  0x07: "D",
  0x08: "E",
  0x09: "F",
  0x0a: "G",
  0x0b: "H",
  0x0c: "I",
  0x0d: "J",
  0x0e: "K",
  0x0f: "L",
  0x10: "M",
  0x11: "N",
  0x12: "O",
  0x13: "P",
  0x14: "Q",
  0x15: "R",
  0x16: "S",
  0x17: "T",
  0x18: "U",
  0x19: "V",
  0x1a: "W",
  0x1b: "X",
  0x1c: "Y",
  0x1d: "Z",
};

export function decodeHIDReport(modifier: number, keycode: number): string | null {
  if (keycode === 0) return null;
  const isShift = (modifier & 0x02) !== 0 || (modifier & 0x20) !== 0;
  const char = isShift ? SHIFT_KEYCODE_TO_CHAR[keycode] : HID_KEYCODE_TO_CHAR[keycode];
  return char ?? null;
}

export interface UseZebraHIDOptions {
  onScan: (sku: string) => void;
  onConnect?: (device: HIDDevice) => void;
  onDisconnect?: () => void;
}

export function useZebraHID({ onScan, onConnect, onDisconnect }: UseZebraHIDOptions) {
  const deviceRef = useRef<HIDDevice | null>(null);
  const bufferRef = useRef("");

  const connect = useCallback(async () => {
    if (!("hid" in navigator)) {
      console.warn("WebHID not supported in this browser");
      return;
    }
    const devices = await navigator.hid.requestDevice({
      filters: [{ vendorId: ZEBRA_VENDOR_ID }],
    });
    const device = devices[0];
    if (!device) return;

    await device.open();
    deviceRef.current = device;
    onConnect?.(device);

    device.addEventListener("inputreport", (e: HIDInputReportEvent) => {
      const view = new DataView(e.data.buffer);
      const modifier = view.getUint8(0);
      const keycode = view.getUint8(2);
      const char = decodeHIDReport(modifier, keycode);
      if (char === "\r") {
        const barcode = bufferRef.current.trim();
        if (barcode) onScan(barcode);
        bufferRef.current = "";
      } else if (char) {
        bufferRef.current += char;
      }
    });

    device.addEventListener("disconnect", () => {
      deviceRef.current = null;
      bufferRef.current = "";
      onDisconnect?.();
    });
  }, [onScan, onConnect, onDisconnect]);

  const disconnect = useCallback(async () => {
    await deviceRef.current?.close();
    deviceRef.current = null;
  }, []);

  useEffect(
    () => () => {
      deviceRef.current?.close();
    },
    [],
  );

  return { connect, disconnect };
}
