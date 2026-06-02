import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { WebhookConfig } from "../WebhookConfig";
import { CONNECTOR_CATALOG, isOutbound, webhookUrl } from "../useIntegrations";

describe("connector catalog helpers", () => {
  it("builds absolute webhook urls and trims trailing slash", () => {
    const stripe = CONNECTOR_CATALOG.find((c) => c.kind === "stripe")!;
    expect(webhookUrl("http://localhost:8080/", stripe)).toBe(
      "http://localhost:8080/webhooks/stripe",
    );
    const klaviyo = CONNECTOR_CATALOG.find((c) => c.kind === "klaviyo")!;
    expect(webhookUrl("http://x", klaviyo)).toBe(""); // outbound, no inbound path
  });

  it("flags outbound connectors", () => {
    expect(isOutbound(CONNECTOR_CATALOG.find((c) => c.kind === "klaviyo")!)).toBe(true);
    expect(isOutbound(CONNECTOR_CATALOG.find((c) => c.kind === "stripe")!)).toBe(false);
  });
});

describe("WebhookConfig", () => {
  it("renders every connector and toggles outbound", () => {
    const onToggle = vi.fn();
    render(<WebhookConfig outboundEnabled={{}} onToggleOutbound={onToggle} />);

    expect(screen.getByText("Stripe Connect")).toBeInTheDocument();
    expect(screen.getByText("coming soon")).toBeInTheDocument(); // NetSuite

    fireEvent.click(screen.getByLabelText("Enable outbound for Klaviyo"));
    expect(onToggle).toHaveBeenCalledWith("klaviyo", true);
  });
});
