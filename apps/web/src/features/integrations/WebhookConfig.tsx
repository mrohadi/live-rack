import { CONNECTOR_CATALOG, isOutbound, webhookUrl, type Connector } from "./useIntegrations";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

interface WebhookConfigProps {
  /** kinds whose outbound push is enabled */
  outboundEnabled: Record<string, boolean>;
  onToggleOutbound: (kind: string, enabled: boolean) => void;
}

function ConnectorRow({
  c,
  enabled,
  onToggle,
}: {
  c: Connector;
  enabled: boolean;
  onToggle: (v: boolean) => void;
}) {
  const url = webhookUrl(API_BASE, c);
  return (
    <div className="flex items-center justify-between gap-3 border-t border-border px-3 py-2">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-foreground">{c.label}</span>
          <span className="rounded bg-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
            {c.direction}
          </span>
          {c.comingSoon && (
            <span className="rounded bg-muted/40 px-1.5 py-0.5 text-[10px] text-muted-foreground">
              coming soon
            </span>
          )}
        </div>
        {url && (
          <code className="mt-0.5 block truncate font-mono text-[11px] text-muted-foreground">
            {url}
          </code>
        )}
      </div>

      {isOutbound(c) ? (
        <label className="flex shrink-0 items-center gap-2 text-xs text-muted-foreground">
          <input
            type="checkbox"
            checked={enabled}
            disabled={c.comingSoon}
            onChange={(e) => onToggle(e.target.checked)}
            aria-label={`Enable outbound for ${c.label}`}
          />
          Push events
        </label>
      ) : (
        <span className="shrink-0 text-[11px] text-muted-foreground">inbound only</span>
      )}
    </div>
  );
}

/** Marketplace connectors with inbound webhook URLs and outbound push toggles. */
export function WebhookConfig({ outboundEnabled, onToggleOutbound }: WebhookConfigProps) {
  return (
    <div
      className="overflow-hidden rounded-lg border border-border bg-surface"
      data-testid="webhook-config"
    >
      {CONNECTOR_CATALOG.map((c) => (
        <ConnectorRow
          key={c.kind}
          c={c}
          enabled={!!outboundEnabled[c.kind]}
          onToggle={(v) => onToggleOutbound(c.kind, v)}
        />
      ))}
    </div>
  );
}
