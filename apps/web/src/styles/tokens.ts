// Design tokens — single source of truth for live-rack.
// tailwind.config.ts maps these via hsl(var(--<token>) / <alpha-value>).
// tokens.css writes them to CSS custom properties for both themes.

export const REQUIRED_COLOR_TOKENS = [
  "color-primary",
  "color-secondary",
  "color-accent",
  "color-surface",
  "color-background",
  "color-foreground",
  "color-muted",
  "color-muted-foreground",
  "color-border",
  "color-destructive",
  "color-success",
  "color-warning",
] as const;

export const REQUIRED_STATIC_TOKENS = [
  "font-sans",
  "font-mono",
  "radius",
  "radius-sm",
  "radius-lg",
] as const;

export type ColorToken = (typeof REQUIRED_COLOR_TOKENS)[number];
export type StaticToken = (typeof REQUIRED_STATIC_TOKENS)[number];

// HSL channel values — no hsl() wrapper, matching tailwind.config.ts pattern:
// colors: { primary: "hsl(var(--color-primary) / <alpha-value>)" }

export const lightTokens: Record<ColorToken, string> = {
  "color-primary": "220 91% 56%",         // #2563eb  blue-600
  "color-secondary": "215 16% 47%",       // #64748b  slate-500
  "color-accent": "220 91% 56%",          // #2563eb  blue-600
  "color-surface": "0 0% 100%",           // #ffffff
  "color-background": "210 40% 98%",      // #f8fafc  slate-50
  "color-foreground": "222 84% 5%",       // #0f172a  slate-950
  "color-muted": "210 40% 96%",           // #f1f5f9  slate-100
  "color-muted-foreground": "215 16% 47%",// #64748b  slate-500
  "color-border": "214 32% 91%",          // #e2e8f0  slate-200
  "color-destructive": "0 72% 51%",       // #dc2626  red-600
  "color-success": "142 71% 45%",         // #16a34a  green-600
  "color-warning": "38 92% 50%",          // #d97706  amber-600
};

export const darkTokens: Record<ColorToken, string> = {
  "color-primary": "217 91% 60%",         // #3b82f6  blue-500
  "color-secondary": "215 20% 65%",       // #94a3b8  slate-400
  "color-accent": "217 91% 60%",          // #3b82f6  blue-500
  "color-surface": "222 47% 11%",         // #1e293b  slate-800
  "color-background": "222 84% 5%",       // #0f172a  slate-950
  "color-foreground": "210 40% 98%",      // #f8fafc  slate-50
  "color-muted": "217 33% 17%",           // #1e293b  slate-800
  "color-muted-foreground": "215 20% 65%",// #94a3b8  slate-400
  "color-border": "217 33% 17%",          // #1e293b  slate-800
  "color-destructive": "0 63% 61%",       // #ef4444  red-500
  "color-success": "142 69% 58%",         // #4ade80  green-400
  "color-warning": "38 92% 62%",          // #fbbf24  amber-400
};

export const staticTokens: Record<StaticToken, string> = {
  "font-sans": "'Inter', system-ui, sans-serif",
  "font-mono": "'JetBrains Mono', 'Fira Code', monospace",
  "radius": "0.5rem",
  "radius-sm": "0.25rem",
  "radius-lg": "0.75rem",
};
