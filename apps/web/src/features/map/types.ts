export type ZoneType = "general" | "frozen" | "returns" | "staging" | "display" | "checkout";
export type ViewMode = "zones" | "heat" | "items";

export interface ZoneConstraints {
  allowedCategories?: string[];
  maxSKUs?: number;
  climate?: string;
}

export interface Zone {
  id: string;
  name: string;
  x: number;
  y: number;
  width: number;
  height: number;
  color: string;
  type: ZoneType;
  items?: number;
  capacity?: number;
  sales?: number;
  misplaced?: number;
  lastScan?: string;
  dwell?: string;
  constraints?: ZoneConstraints;
}

/* Partial zone update emitted by the editor on drag/resize */
export interface ZoneUpdate {
  id: string;
  x?: number;
  y?: number;
  width?: number;
  height?: number;
}

export interface ZoneCanvasProps {
  zones: Zone[];
  // Multi-select aware. Empty array = nothing selected.
  selectedIds: string[];
  onSelect: (ids: string[]) => void;
  /** Optional — when present, canvas becomes editable. Receives one or more changes. */
  onChange?: (updates: ZoneUpdate[]) => void;
  /** Snap step in CSS px. Default 10. */
  gridSize?: number;
  /** Show grid background. Default true when onChange present. */
  showGrid?: boolean;
  viewMode?: ViewMode;
}
