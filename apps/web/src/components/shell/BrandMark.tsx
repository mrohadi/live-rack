interface BrandMarkProps {
  accent?: string;
}

export function BrandMark({ accent = "#2563eb" }: BrandMarkProps) {
  return (
    <svg className="brand-mark" viewBox="0 0 32 32">
      <rect x="2" y="2" width="28" height="28" rx="7" fill={accent} />
      <rect x="6" y="9" width="20" height="3.2" rx="1" fill="white" opacity="0.95" />
      <rect x="6" y="14.4" width="13" height="3.2" rx="1" fill="white" opacity="0.7" />
      <rect x="6" y="19.8" width="20" height="3.2" rx="1" fill="white" opacity="0.95" />
      <circle cx="22" cy="16" r="1.6" fill="white" />
    </svg>
  );
}
