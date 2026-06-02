import { useEffect, useRef, useState } from "react";

interface Props {
  /** ISO date string YYYY-MM-DD or empty string. */
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

const DAYS = ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"];
const MONTHS = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

function parseDate(iso: string): Date | null {
  if (!iso) return null;
  const d = new Date(iso + "T00:00:00");
  return isNaN(d.getTime()) ? null : d;
}

function toISO(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function formatDisplay(iso: string): string {
  const d = parseDate(iso);
  if (!d) return "";
  return d.toLocaleDateString(undefined, { day: "numeric", month: "short", year: "numeric" });
}

export function DatePicker({
  value,
  onChange,
  placeholder = "Pick a date…",
  disabled = false,
  className = "",
}: Props) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const selected = parseDate(value);
  const [open, setOpen] = useState(false);
  const [cursor, setCursor] = useState<Date>(
    selected
      ? new Date(selected.getFullYear(), selected.getMonth(), 1)
      : new Date(today.getFullYear(), today.getMonth(), 1),
  );
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  const prevMonth = () => setCursor((c) => new Date(c.getFullYear(), c.getMonth() - 1, 1));
  const nextMonth = () => setCursor((c) => new Date(c.getFullYear(), c.getMonth() + 1, 1));

  /** Build the grid: weeks × days, with leading/trailing nulls. */
  function buildGrid(): (Date | null)[][] {
    const year = cursor.getFullYear();
    const month = cursor.getMonth();
    const first = new Date(year, month, 1);
    const last = new Date(year, month + 1, 0);
    const cells: (Date | null)[] = Array(first.getDay()).fill(null);
    for (let d = 1; d <= last.getDate(); d++) cells.push(new Date(year, month, d));
    while (cells.length % 7 !== 0) cells.push(null);
    const weeks: (Date | null)[][] = [];
    for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7));
    return weeks;
  }

  const pick = (d: Date) => {
    onChange(toISO(d));
    setOpen(false);
  };

  const clear = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange("");
  };

  return (
    <div ref={ref} className={`relative ${className}`}>
      {/* Trigger */}
      <button
        type="button"
        disabled={disabled}
        onClick={() => setOpen((o) => !o)}
        className={`flex w-full items-center justify-between gap-2 rounded-lg border px-3 py-2 text-sm transition ${
          open
            ? "border-primary ring-2 ring-primary/20"
            : "border-border bg-background hover:border-primary/50"
        } disabled:cursor-not-allowed disabled:opacity-50`}
      >
        <span className="flex items-center gap-2">
          <CalendarIcon />
          {value ? (
            <span className="text-foreground">{formatDisplay(value)}</span>
          ) : (
            <span className="text-muted-foreground">{placeholder}</span>
          )}
        </span>
        {value ? (
          <span
            onClick={clear}
            className="rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
            role="button"
            aria-label="Clear date"
          >
            <XIcon />
          </span>
        ) : (
          <ChevronIcon open={open} />
        )}
      </button>

      {/* Calendar panel */}
      {open && (
        <div className="absolute left-0 right-0 top-full z-50 mt-1 rounded-lg border border-border bg-surface p-3 shadow-lg">
          {/* Month header */}
          <div className="mb-3 flex items-center justify-between">
            <button
              type="button"
              onClick={prevMonth}
              className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
              aria-label="Previous month"
            >
              <ChevronLeftIcon />
            </button>
            <span className="text-sm font-semibold text-foreground">
              {MONTHS[cursor.getMonth()]} {cursor.getFullYear()}
            </span>
            <button
              type="button"
              onClick={nextMonth}
              className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
              aria-label="Next month"
            >
              <ChevronRightIcon />
            </button>
          </div>

          {/* Day-of-week header */}
          <div className="mb-1 grid grid-cols-7 text-center">
            {DAYS.map((d) => (
              <span key={d} className="text-[10px] font-semibold uppercase text-muted-foreground">
                {d}
              </span>
            ))}
          </div>

          {/* Calendar grid */}
          {buildGrid().map((week, wi) => (
            <div key={wi} className="grid grid-cols-7">
              {week.map((day, di) => {
                if (!day) return <span key={di} />;
                const iso = toISO(day);
                const isSelected = iso === value;
                const isToday = iso === toISO(today);
                const isPast = day < today;
                return (
                  <button
                    key={di}
                    type="button"
                    onClick={() => pick(day)}
                    disabled={isPast}
                    className={`mx-auto flex h-8 w-8 items-center justify-center rounded-full text-xs transition ${
                      isSelected
                        ? "bg-primary font-semibold text-white"
                        : isToday
                          ? "border border-primary text-primary hover:bg-primary/10"
                          : isPast
                            ? "cursor-not-allowed text-muted-foreground/40"
                            : "text-foreground hover:bg-muted"
                    }`}
                  >
                    {day.getDate()}
                  </button>
                );
              })}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function CalendarIcon() {
  return (
    <svg className="h-4 w-4 shrink-0 text-muted-foreground" viewBox="0 0 20 20" fill="currentColor">
      <path
        fillRule="evenodd"
        d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function XIcon() {
  return (
    <svg className="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
      <path
        fillRule="evenodd"
        d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function ChevronIcon({ open }: { open: boolean }) {
  return (
    <svg
      className={`h-4 w-4 shrink-0 text-muted-foreground transition-transform ${open ? "rotate-180" : ""}`}
      viewBox="0 0 20 20"
      fill="currentColor"
    >
      <path
        fillRule="evenodd"
        d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function ChevronLeftIcon() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
      <path
        fillRule="evenodd"
        d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function ChevronRightIcon() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
      <path
        fillRule="evenodd"
        d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
        clipRule="evenodd"
      />
    </svg>
  );
}
