import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { SearchResult } from "./types";
import { MIN_QUERY_LENGTH, useSearch } from "./useSearch";

/** Where selecting a hit of each kind navigates. */
const KIND_ROUTE: Record<SearchResult["kind"], string> = {
  item: "/inventory",
  zone: "/map",
};

const KIND_LABEL: Record<SearchResult["kind"], string> = {
  item: "Item",
  zone: "Zone",
};

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
}

export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [active, setActive] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const { data: results = [], isFetching } = useSearch(query);

  // Reset and focus whenever the palette opens.
  useEffect(() => {
    if (!open) return;
    setQuery("");
    setActive(0);
    inputRef.current?.focus();
  }, [open]);

  // Keep the highlighted row in range as results change.
  useEffect(() => {
    setActive(0);
  }, [results]);

  const select = useCallback(
    (r: SearchResult | undefined) => {
      if (!r) return;
      onClose();
      navigate(KIND_ROUTE[r.kind]);
    },
    [navigate, onClose],
  );

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      onClose();
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      setActive((i) => Math.min(i + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter") {
      e.preventDefault();
      select(results[active]);
    }
  };

  if (!open) return null;

  const showEmpty = query.trim().length >= MIN_QUERY_LENGTH && !isFetching && results.length === 0;

  return (
    <div className="cmdk-overlay" onClick={onClose} role="presentation">
      <div
        className="cmdk-panel"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Search"
      >
        <input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={onKeyDown}
          placeholder="Search SKUs, items, zones…"
          className="cmdk-input"
        />
        <ul className="cmdk-list">
          {results.map((r, i) => (
            <li key={`${r.kind}-${r.id}`}>
              <button
                type="button"
                data-testid="search-result"
                onClick={() => select(r)}
                onMouseEnter={() => setActive(i)}
                className={`cmdk-item${i === active ? " active" : ""}`}
              >
                <span className="cmdk-label">
                  {r.label}
                  {r.sublabel && <span className="cmdk-sku">{r.sublabel}</span>}
                </span>
                <span className="cmdk-kind">{KIND_LABEL[r.kind]}</span>
              </button>
            </li>
          ))}
          {showEmpty && <li className="cmdk-empty">No matches.</li>}
        </ul>
      </div>
    </div>
  );
}
