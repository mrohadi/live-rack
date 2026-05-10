import Konva from "konva";
import { useEffect, useRef } from "react";
import type { ZoneCanvasProps } from "../types";

export function KonvaZoneCanvas({ zones, selectedId, onSelect }: ZoneCanvasProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const stageRef = useRef<Konva.Stage | null>(null);
  const layerRef = useRef<Konva.Layer | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    const stage = new Konva.Stage({
      container: containerRef.current,
      width: containerRef.current.offsetWidth || 1200,
      height: containerRef.current.offsetHeight || 800,
    });

    const layer = new Konva.Layer();
    stage.add(layer);

    zones.forEach((zone) => {
      const rect = new Konva.Rect({
        id: zone.id,
        x: zone.x,
        y: zone.y,
        width: zone.width,
        height: zone.height,
        fill: zone.color,
        stroke: zone.id === selectedId ? "#fff" : "transparent",
        strokeWidth: zone.id === selectedId ? 2 : 0,
        cornerRadius: 4,
      });

      rect.on("click", () => onSelect(zone.id));
      layer.add(rect);
    });

    layer.draw();

    stageRef.current = stage;
    layerRef.current = layer;

    return () => {
      stage.destroy();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [zones]);

  useEffect(() => {
    const layer = layerRef.current;
    if (!layer) return;

    layer.find("Rect").forEach((node) => {
      const rect = node as Konva.Rect;
      const isSelected = rect.id() === selectedId;
      rect.stroke(isSelected ? "#fff" : "transparent");
      rect.strokeWidth(isSelected ? 2 : 0);
    });

    layer.batchDraw();
  }, [selectedId]);

  return <div ref={containerRef} style={{ width: "100%", height: "100%" }} aria-label="Zone map" />;
}
