import { statusTone, useIntegrations, useWebhookLog } from "./useIntegrations";

export function IntegrationsPage() {
  const { data: integrations = [], isLoading: loadingIntegrations } = useIntegrations();
  const { data: events = [], isLoading: loadingLog } = useWebhookLog();

  return (
    <div className="flex h-full flex-col">
      <header className="border-b border-border px-4 py-3">
        <h1 className="text-lg font-semibold text-foreground">Integrations</h1>
        <p className="text-xs text-muted-foreground">POS connectors · inbound webhook log</p>
      </header>

      <div className="flex-1 space-y-6 overflow-auto p-4">
        <section>
          <h2 className="mb-2 text-sm font-semibold text-foreground">Connectors</h2>
          {loadingIntegrations ? (
            <div className="text-sm text-muted-foreground">Loading…</div>
          ) : integrations.length === 0 ? (
            <div className="text-sm text-muted-foreground">No integrations connected.</div>
          ) : (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {integrations.map((i) => (
                <div key={i.id} className="rounded-lg border border-border bg-surface p-3">
                  <div className="flex items-center justify-between">
                    <span className="font-medium capitalize text-foreground">{i.kind}</span>
                    <span className={`rounded-full px-2 py-0.5 text-xs ${statusTone(i.status)}`}>
                      {i.status}
                    </span>
                  </div>
                  <div className="mt-1 truncate font-mono text-[11px] text-muted-foreground">
                    {i.external_id || "—"}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        <section>
          <h2 className="mb-2 text-sm font-semibold text-foreground">Webhook event log</h2>
          {loadingLog ? (
            <div className="text-sm text-muted-foreground">Loading…</div>
          ) : events.length === 0 ? (
            <div className="text-sm text-muted-foreground">No webhook deliveries yet.</div>
          ) : (
            <div className="overflow-hidden rounded-lg border border-border">
              <table className="w-full text-sm">
                <thead className="bg-muted/30 text-left text-xs uppercase tracking-wide text-muted-foreground">
                  <tr>
                    <th className="px-3 py-2">Provider</th>
                    <th className="px-3 py-2">Topic</th>
                    <th className="px-3 py-2">Event ID</th>
                    <th className="px-3 py-2">Status</th>
                    <th className="px-3 py-2">Received</th>
                  </tr>
                </thead>
                <tbody>
                  {events.map((e) => (
                    <tr key={e.id} className="border-t border-border">
                      <td className="px-3 py-2 capitalize text-foreground">{e.provider}</td>
                      <td className="px-3 py-2 text-muted-foreground">{e.topic || "—"}</td>
                      <td className="px-3 py-2 font-mono text-[11px] text-muted-foreground">
                        {e.event_id}
                      </td>
                      <td className="px-3 py-2">
                        <span className={`rounded-full px-2 py-0.5 text-xs ${statusTone(e.status)}`}>
                          {e.status}
                        </span>
                      </td>
                      <td className="px-3 py-2 font-mono text-[11px] text-muted-foreground">
                        {new Date(e.received_at).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
