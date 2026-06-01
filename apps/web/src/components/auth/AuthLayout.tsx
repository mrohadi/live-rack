import type { ReactNode } from "react";

function LogoMark() {
  return (
    <svg viewBox="0 0 32 32" className="h-9 w-9" aria-hidden="true">
      <rect x="2" y="2" width="28" height="28" rx="7" fill="white" />
      <rect x="6" y="9" width="20" height="3.2" rx="1" fill="hsl(220 91% 56%)" opacity="0.95" />
      <rect x="6" y="14.4" width="13" height="3.2" rx="1" fill="hsl(220 91% 56%)" opacity="0.6" />
      <rect x="6" y="19.8" width="20" height="3.2" rx="1" fill="hsl(220 91% 56%)" opacity="0.95" />
      <circle cx="22" cy="16" r="1.6" fill="hsl(220 91% 56%)" />
    </svg>
  );
}

const HIGHLIGHTS = [
  "Live zone map & barcode scanning",
  "Pipelines, tasks & inventory in one rack",
  "Analytics, signals & smart recommendations",
];

interface AuthLayoutProps {
  title: string;
  subtitle?: string;
  children: ReactNode;
}

/** Two-pane auth shell: brand story on the left, form on the right.
 *  Matches the live-rack palette (primary #2563eb, Inter). */
export function AuthLayout({ title, subtitle, children }: AuthLayoutProps) {
  return (
    <div className="flex min-h-screen bg-background font-sans text-foreground">
      {/* Brand panel — hidden on small screens. */}
      <aside className="relative hidden w-1/2 flex-col justify-between overflow-hidden bg-primary p-10 text-white lg:flex">
        <div
          className="pointer-events-none absolute inset-0 opacity-20"
          style={{
            backgroundImage:
              "radial-gradient(circle at 20% 20%, white 0, transparent 40%), radial-gradient(circle at 80% 60%, white 0, transparent 35%)",
          }}
        />
        <div className="relative flex items-center gap-2.5">
          <LogoMark />
          <span className="text-lg font-semibold tracking-tight">live-rack</span>
        </div>

        <div className="relative space-y-6">
          <h2 className="max-w-sm text-3xl font-semibold leading-tight">
            Run your floor in real time.
          </h2>
          <ul className="space-y-3">
            {HIGHLIGHTS.map((h) => (
              <li key={h} className="flex items-center gap-3 text-sm text-white/90">
                <span className="flex h-5 w-5 items-center justify-center rounded-full bg-white/20 text-xs">
                  ✓
                </span>
                {h}
              </li>
            ))}
          </ul>
        </div>

        <p className="relative text-xs text-white/60">
          © live-rack — warehouse operations platform
        </p>
      </aside>

      {/* Form panel. */}
      <main className="flex w-full flex-col items-center justify-center px-6 py-12 lg:w-1/2">
        <div className="w-full max-w-sm">
          <div className="mb-8 flex items-center gap-2 lg:hidden">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary">
              <span className="text-sm font-bold text-white">lr</span>
            </div>
            <span className="text-lg font-semibold tracking-tight">live-rack</span>
          </div>

          <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
          {subtitle && <p className="mt-1.5 text-sm text-muted-foreground">{subtitle}</p>}

          <div className="mt-8">{children}</div>
        </div>
      </main>
    </div>
  );
}
