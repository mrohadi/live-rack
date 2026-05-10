export type ZoneType = "general" | "frozen" | "returns" | "staging" | "display" | "checkout";

export interface Zone {
  id: string;
  name: string;
  x: number;
  y: number;
  width: number;
  height: number;
  color: string;
  type: ZoneType;
}

export interface ZoneCanvasProps {
  zones: Zone[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}
