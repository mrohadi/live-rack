import { BrandMark } from "./shell/BrandMark";

interface LoadingScreenProps {
  label?: string;
}

// Branded full-screen loader: pulsing logo over a rotating accent ring.
export function LoadingScreen({ label = "Loading…" }: LoadingScreenProps) {
  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 18,
        height: "100vh",
        background: "var(--bg)",
        color: "var(--text-3)",
      }}
    >
      <div className="lr-loader">
        <span className="lr-loader-ring" />
        <span className="lr-loader-mark">
          <BrandMark />
        </span>
      </div>
      <span style={{ fontSize: 13, fontFamily: "var(--mono)", letterSpacing: "0.02em" }}>
        {label}
      </span>
    </div>
  );
}
