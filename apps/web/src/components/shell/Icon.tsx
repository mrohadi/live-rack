import type { ReactNode } from "react";

interface IconProps {
  d: string | ReactNode;
  size?: number;
  stroke?: number;
  className?: string;
}

export function Icon({ d, size = 16, stroke = 1.6, className }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      width={size}
      height={size}
      fill="none"
      stroke="currentColor"
      strokeWidth={stroke}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      {typeof d === "string" ? <path d={d} /> : d}
    </svg>
  );
}

export const Icons = {
  dash: "M3 13l9-9 9 9M5 11v9h5v-6h4v6h5v-9",
  map: (
    <>
      <path d="M9 4l-6 2v14l6-2 6 2 6-2V4l-6 2-6-2z" />
      <path d="M9 4v14M15 6v14" />
    </>
  ),
  scan: (
    <>
      <path d="M4 7V4h3M20 7V4h-3M4 17v3h3M20 17v3h-3" />
      <path d="M4 12h16" />
    </>
  ),
  box: (
    <>
      <path d="M21 8l-9-5-9 5 9 5 9-5z" />
      <path d="M3 8v8l9 5 9-5V8" />
      <path d="M12 13v8" />
    </>
  ),
  pipe: (
    <>
      <path d="M3 6h6v6H3zM15 6h6v6h-6zM9 18h6" />
      <path d="M9 9h6M12 12v6" />
    </>
  ),
  task: (
    <>
      <path d="M9 11l3 3L22 4" />
      <path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" />
    </>
  ),
  pick: (
    <>
      <path d="M9 4H7a2 2 0 00-2 2v13a2 2 0 002 2h10a2 2 0 002-2V6a2 2 0 00-2-2h-2" />
      <path d="M9 3a1 1 0 011-1h4a1 1 0 011 1v2H9V3z" />
      <path d="M9 11l1.5 1.5L13 10M9 16l1.5 1.5L13 15" />
    </>
  ),
  wave: (
    <>
      <path d="M2 7h7M2 12h12M2 17h9" />
      <path d="M16 6l4 4-4 4" />
    </>
  ),
  chart: (
    <>
      <path d="M3 3v18h18" />
      <path d="M7 15l4-4 3 3 5-7" />
    </>
  ),
  plug: (
    <>
      <path d="M9 2v4M15 2v4M7 6h10v6a5 5 0 01-10 0V6zM12 17v5" />
    </>
  ),
  bell: (
    <>
      <path d="M6 8a6 6 0 1112 0c0 7 3 8 3 8H3s3-1 3-8" />
      <path d="M10 21a2 2 0 004 0" />
    </>
  ),
  search: (
    <>
      <circle cx="11" cy="11" r="7" />
      <path d="M21 21l-4.3-4.3" />
    </>
  ),
  menu: (
    <>
      <path d="M3 6h18M3 12h18M3 18h18" />
    </>
  ),
  cog: (
    <>
      <circle cx="12" cy="12" r="3" />
      <path d="M19.4 15a1.7 1.7 0 00.3 1.8l.1.1a2 2 0 11-2.8 2.8l-.1-.1a1.7 1.7 0 00-1.8-.3 1.7 1.7 0 00-1 1.5V21a2 2 0 11-4 0v-.1a1.7 1.7 0 00-1-1.5 1.7 1.7 0 00-1.8.3l-.1.1a2 2 0 11-2.8-2.8l.1-.1a1.7 1.7 0 00.3-1.8 1.7 1.7 0 00-1.5-1H3a2 2 0 110-4h.1a1.7 1.7 0 001.5-1 1.7 1.7 0 00-.3-1.8l-.1-.1a2 2 0 112.8-2.8l.1.1a1.7 1.7 0 001.8.3H9a1.7 1.7 0 001-1.5V3a2 2 0 114 0v.1a1.7 1.7 0 001 1.5 1.7 1.7 0 001.8-.3l.1-.1a2 2 0 112.8 2.8l-.1.1a1.7 1.7 0 00-.3 1.8V9a1.7 1.7 0 001.5 1H21a2 2 0 110 4h-.1a1.7 1.7 0 00-1.5 1z" />
    </>
  ),
  user: (
    <>
      <circle cx="12" cy="8" r="4" />
      <path d="M4 21a8 8 0 0116 0" />
    </>
  ),
} as const;
