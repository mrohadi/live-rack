import type { Zone } from "./types";

interface Props {
  zone: Zone | null;
}

export function ZoneDetailSidebar({ zone }: Props) {
  if (!zone) {
    return (
      <div
        style={{
          width: 280,
          borderLeft: "1px solid #1f2937",
          background: "#0f172a",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: "#475569",
          fontSize: 13,
        }}
      >
        Select a zone to inspect
      </div>
    );
  }

  const fill =
    zone.items != null && zone.capacity ? Math.round((zone.items / zone.capacity) * 100) : 0;

  const { constraints } = zone;

  return (
    <div
      style={{
        width: 280,
        borderLeft: "1px solid #1f2937",
        background: "#0f172a",
        display: "flex",
        flexDirection: "column",
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: "12px 16px",
          borderBottom: "1px solid #1f2937",
          display: "flex",
          alignItems: "flex-start",
        }}
      >
        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: 600, fontSize: 15, color: "#f1f5f9" }}>{zone.name}</div>
          <div style={{ fontSize: 12, color: "#64748b", marginTop: 2 }}>
            {zone.type} · zone {zone.id}
          </div>
        </div>
        <button
          aria-label="more"
          style={{
            background: "transparent",
            border: "none",
            cursor: "pointer",
            color: "#64748b",
            fontSize: 18,
            lineHeight: 1,
          }}
        >
          ⋯
        </button>
      </div>

      {/* Body */}
      <div style={{ flex: 1, padding: "10px 16px", overflowY: "auto" }}>
        <KV label="Capacity" value={`${zone.items ?? 0} / ${zone.capacity ?? 0}`} />
        <KV label="Fill" value={`${fill}%`} />
        {zone.sales != null && <KV label="Sales today" value={`$${zone.sales.toLocaleString()}`} />}
        {zone.dwell && <KV label="Avg. dwell" value={zone.dwell} />}
        {zone.misplaced != null && <KV label="Misplaced" value={String(zone.misplaced)} />}
        {zone.lastScan && <KV label="Last scan" value={zone.lastScan} mono />}

        {constraints && (
          <>
            <div
              style={{
                marginTop: 14,
                fontSize: 11,
                fontWeight: 700,
                color: "#64748b",
                letterSpacing: "0.08em",
              }}
            >
              CONSTRAINTS
            </div>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 6, marginTop: 8 }}>
              {constraints.allowedCategories?.map((cat) => (
                <Chip key={cat} accent>
                  {cat} only
                </Chip>
              ))}
              {constraints.maxSKUs != null && <Chip>Max {constraints.maxSKUs} SKUs</Chip>}
              {constraints.climate && <Chip>Climate: {constraints.climate}</Chip>}
            </div>
          </>
        )}
      </div>

      {/* Footer */}
      <div
        style={{
          padding: "10px 16px",
          borderTop: "1px solid #1f2937",
          display: "flex",
          gap: 8,
        }}
      >
        <button
          style={{
            flex: 1,
            padding: "6px 0",
            borderRadius: 6,
            border: "1px solid #334155",
            background: "transparent",
            color: "#cbd5e1",
            cursor: "pointer",
            fontSize: 13,
          }}
        >
          Open zone
        </button>
        <button
          style={{
            flex: 1,
            padding: "6px 0",
            borderRadius: 6,
            border: "none",
            background: "#6366f1",
            color: "#fff",
            cursor: "pointer",
            fontSize: 13,
            fontWeight: 600,
          }}
        >
          Assign task
        </button>
      </div>
    </div>
  );
}

function KV({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div
      style={{ display: "flex", justifyContent: "space-between", padding: "4px 0", fontSize: 13 }}
    >
      <span style={{ color: "#64748b" }}>{label}</span>
      <span style={{ color: "#f1f5f9", fontFamily: mono ? "var(--mono, monospace)" : undefined }}>
        {value}
      </span>
    </div>
  );
}

function Chip({ children, accent }: { children: React.ReactNode; accent?: boolean }) {
  return (
    <span
      style={{
        padding: "2px 8px",
        borderRadius: 999,
        fontSize: 12,
        background: accent ? "#312e81" : "#1e293b",
        color: accent ? "#a5b4fc" : "#94a3b8",
        border: `1px solid ${accent ? "#4338ca" : "#334155"}`,
      }}
    >
      {children}
    </span>
  );
}
