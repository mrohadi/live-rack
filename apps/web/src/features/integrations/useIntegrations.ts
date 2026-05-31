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
