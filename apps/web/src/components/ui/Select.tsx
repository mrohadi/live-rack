import { useEffect, useRef, useState } from "react";

export interface SelectOption {
  value: string;
  label: string;
  /** Short text rendered as a coloured badge beside the label. */
  badge?: string;
  badgeClass?: string;
  /** One or two initials shown in an avatar circle. */
  avatar?: string;
  avatarUrl?: string;
  /** Dim line below the label. */
  sub?: string;
}

interface Props {
  value: string;
  onChange: (v: string) => void;
  options: SelectOption[];
  placeholder?: string;
  searchable?: boolean;
  disabled?: boolean;
  /** Extra class on the trigger button. */
  className?: string;
}

export function Select({
  value,
  onChange,
  options,
  placeholder = "Select…",
  searchable = false,
  disabled = false,
  className = "",
}: Props) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const ref = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  // Close on outside click.
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  // Focus search input when opened.
  useEffect(() => {
    if (open && searchable) setTimeout(() => searchRef.current?.focus(), 0);
  }, [open, searchable]);

  const selected = options.find((o) => o.value === value);

  const filtered =
    searchable && query
      ? options.filter(
          (o) =>
            o.label.toLowerCase().includes(query.toLowerCase()) ||
            (o.sub ?? "").toLowerCase().includes(query.toLowerCase()),
        )
      : options;

  const pick = (v: string) => {
    onChange(v);
    setOpen(false);
    setQuery("");
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
        <span className="flex min-w-0 items-center gap-2">
          {selected ? (
            <OptionDisplay opt={selected} />
          ) : (
            <span className="text-muted-foreground">{placeholder}</span>
          )}
        </span>
        <ChevronIcon open={open} />
      </button>

      {/* Dropdown panel */}
      {open && (
        <div className="absolute left-0 right-0 top-full z-50 mt-1 overflow-hidden rounded-lg border border-border bg-surface shadow-lg">
          {searchable && (
            <div className="border-b border-border px-3 py-2">
              <input
                ref={searchRef}
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search…"
                className="w-full bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
              />
            </div>
          )}
          <ul className="max-h-52 overflow-y-auto py-1" role="listbox">
            {filtered.length === 0 && (
              <li className="px-3 py-2 text-sm text-muted-foreground">No results</li>
            )}
            {filtered.map((opt) => (
              <li
                key={opt.value}
                role="option"
                aria-selected={opt.value === value}
                onClick={() => pick(opt.value)}
                className={`flex cursor-pointer items-center gap-2 px-3 py-2 text-sm transition hover:bg-muted ${
                  opt.value === value ? "bg-primary/8 font-medium text-primary" : "text-foreground"
                }`}
              >
                <OptionDisplay opt={opt} />
                {opt.value === value && (
                  <span className="ml-auto text-primary">
                    <CheckIcon />
                  </span>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

function OptionDisplay({ opt }: { opt: SelectOption }) {
  return (
    <>
      {/* Avatar */}
      {opt.avatar && (
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/20 text-[10px] font-semibold text-primary">
          {opt.avatar}
        </span>
      )}
      {/* Label + sub */}
      <span className="min-w-0 flex-1 truncate">
        {opt.label}
        {opt.sub && <span className="ml-1 text-xs text-muted-foreground">{opt.sub}</span>}
      </span>
      {/* Badge */}
      {opt.badge && (
        <span
          className={`shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${opt.badgeClass ?? "bg-muted text-muted-foreground"}`}
        >
          {opt.badge}
        </span>
      )}
    </>
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

function CheckIcon() {
  return (
    <svg className="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
      <path
        fillRule="evenodd"
        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
        clipRule="evenodd"
      />
    </svg>
  );
}
