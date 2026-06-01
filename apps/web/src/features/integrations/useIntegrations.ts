import { useQuery } from "@tanstack/react-query";
import { useApi } from "../../lib/api";

export interface Integration {
  id: string;
  kind: string;
  status: string;
  external_id: string;
}

export interface WebhookEvent {
  id: string;
  provider: string;
  event_id: string;
  topic: string;
  status: string;
  received_at: string;
}

/** Tailwind classes for a connector/webhook status badge. Pure. */
export function statusTone(status: string): string {
  switch (status) {
    case "connected":
    case "processed":
      return "bg-success/15 text-success";
    case "error":
    case "rejected":
      return "bg-destructive/15 text-destructive";
    default:
      return "bg-muted/40 text-muted-foreground";
  }
}

export const integrationKeys = {
  list: ["integrations", "list"] as const,
  webhooks: ["integrations", "webhooks"] as const,
};

export type ConnectorDirection = "inbound" | "outbound";

export interface Connector {
  kind: string;
  label: string;
  direction: ConnectorDirection;
  /** Inbound webhook path (relative to the API base), if any. */
  webhookPath?: string;
  comingSoon?: boolean;
}

/** Marketplace catalog of supported connectors. */
export const CONNECTOR_CATALOG: Connector[] = [
  { kind: "shopify", label: "Shopify", direction: "inbound", webhookPath: "/webhooks/shopify" },
  { kind: "square", label: "Square", direction: "inbound", webhookPath: "/webhooks/square" },
  {
    kind: "stripe",
    label: "Stripe Connect",
    direction: "inbound",
    webhookPath: "/webhooks/stripe",
  },
  { kind: "shippo", label: "Shippo", direction: "inbound", webhookPath: "/webhooks/shippo" },
  { kind: "klaviyo", label: "Klaviyo", direction: "outbound" },
  { kind: "netsuite", label: "NetSuite", direction: "inbound", comingSoon: true },
];

/** Absolute inbound webhook URL for a connector, or "" if it has none. Pure. */
export function webhookUrl(apiBase: string, c: Connector): string {
  if (!c.webhookPath) return "";
  return `${apiBase.replace(/\/$/, "")}${c.webhookPath}`;
}

/** Connectors that push data out to a third party. Pure. */
export function isOutbound(c: Connector): boolean {
  return c.direction === "outbound";
}

export function useIntegrations() {
  const { get } = useApi();
  return useQuery({
    queryKey: integrationKeys.list,
    queryFn: () => get<Integration[]>("/api/v1/integrations"),
  });
}

export function useWebhookLog() {
  const { get } = useApi();
  return useQuery({
    queryKey: integrationKeys.webhooks,
    queryFn: () => get<WebhookEvent[]>("/api/v1/integrations/webhooks"),
  });
}
