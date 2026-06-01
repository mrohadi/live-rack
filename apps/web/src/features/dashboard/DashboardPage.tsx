import { formatCents, sparkPoints, useSalesSummary } from "./useSales";

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-surface p-4">
      <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </div>
      <div className="mt-1 text-2xl font-semibold text-foreground">{value}</div>
    </div>
  );
}

export function DashboardPage() {
  const { data, isLoading } = useSalesSummary();

  return (
    <div className="flex h-full flex-col">
      <header className="border-b border-border px-4 py-3">
        <h1 className="text-lg font-semibold text-foreground">Overview</h1>
        <p className="text-xs text-muted-foreground">Sales · last 24 hours</p>
      </header>

      <div className="flex-1 overflow-auto p-4">
        {isLoading || !data ? (
          <div className="text-sm text-muted-foreground">Loading sales…</div>
        ) : (
          <>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <StatCard label="Revenue" value={formatCents(data.revenue_cents)} />
              <StatCard label="Units sold" value={String(data.units)} />
              <StatCard label="Orders" value={String(data.orders)} />
            </div>

            <div className="mt-4 rounded-lg border border-border bg-surface p-4">
              <div className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Revenue · 7 days
              </div>
              <svg
                role="img"
                aria-label="7-day revenue sparkline"
                viewBox="0 0 200 40"
                preserveAspectRatio="none"
                className="h-12 w-full"
              >
                <polyline
                  points={sparkPoints(data.spark, 200, 40)}
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  className="text-primary"
                />
              </svg>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
