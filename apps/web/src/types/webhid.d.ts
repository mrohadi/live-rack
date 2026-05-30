interface HIDDevice extends EventTarget {
  readonly productName: string;
  readonly vendorId: number;
  readonly productId: number;
  open(): Promise<void>;
  close(): Promise<void>;
  addEventListener(type: "inputreport", listener: (e: HIDInputReportEvent) => void): void;
  addEventListener(type: "disconnect", listener: () => void): void;
}

interface HIDInputReportEvent extends Event {
  readonly data: DataView;
  readonly device: HIDDevice;
  readonly reportId: number;
}

interface HIDDeviceFilter {
  vendorId?: number;
  productId?: number;
  usagePage?: number;
  usage?: number;
}

interface HIDDeviceRequestOptions {
  filters: HIDDeviceFilter[];
}

interface HID extends EventTarget {
  requestDevice(options: HIDDeviceRequestOptions): Promise<HIDDevice[]>;
  getDevices(): Promise<HIDDevice[]>;
}

interface Navigator {
  readonly hid: HID;
}
