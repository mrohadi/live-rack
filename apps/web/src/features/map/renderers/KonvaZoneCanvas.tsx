import Konva from "konva";
import { useEffect, useRef } from "react";
import type { ZoneCanvasProps, ZoneUpdate } from "../types";

export function KonvaZoneCanvas({
  zones,
  selectedIds,
  onSelect,
  onChange,
  gridSize = 10,
  viewMode = "zones",
}: ZoneCanvasProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const stageRef = useRef<Konva.Stage | null>(null);
  const layerRef = useRef<Konva.Layer | null>(null);
  const selectedIdsRef = useRef<string[]>(selectedIds);

  const editable = Boolean(onChange);

  useEffect(() => {
    if (!containerRef.current) return;

    const stage = new Konva.Stage({
      container: containerRef.current,
      width: containerRef.current.offsetWidth || 1200,
      height: containerRef.current.offsetHeight || 800,
    });

    // Grid layer — behind zones, non-interactive
    const gridLayer = new Konva.Layer({ listening: false });
    const w = stage.width();
    const h = stage.height();
    for (let x = 0; x <= w; x += gridSize) {
      gridLayer.add(
        new Konva.Line({
          points: [x, 0, x, h],
          stroke: "#1f2937",
          strokeWidth: 1,
          opacity: 0.3,
        }),
      );
    }
    for (let y = 0; y <= h; y += gridSize) {
      gridLayer.add(
        new Konva.Line({
          points: [0, y, w, y],
          stroke: "#1f2937",
          strokeWidth: 1,
          opacity: 0.3,
        }),
      );
    }
    stage.add(gridLayer);
    gridLayer.moveToBottom();

    const layer = new Konva.Layer();
    stage.add(layer);

    stage.on("click tap", (e) => {
      if (e.target === stage) onSelect([]);
    });

    zones.forEach((zone) => {
      const rect = new Konva.Rect({
        id: zone.id,
        x: zone.x,
        y: zone.y,
        width: zone.width,
        height: zone.height,
        fill: zone.color,
        stroke: selectedIds.includes(zone.id) ? "#fff" : "transparent",
        strokeWidth: selectedIds.includes(zone.id) ? 2 : 0,
        cornerRadius: 4,
        draggable: editable,
      });

      rect.on("click tap", (e) => {
        const evt = e.evt as MouseEvent | TouchEvent;
        const additive = "shiftKey" in evt && evt.shiftKey;
        const current = selectedIdsRef.current;
        if (additive) {
          const next = current.includes(zone.id)
            ? current.filter((id) => id !== zone.id)
            : [...current, zone.id];
          onSelect(next);
        } else {
          onSelect([zone.id]);
        }
      });

      if (editable) {
        rect.on("dragend", () => {
          const update: ZoneUpdate = {
            id: zone.id,
            x: Math.round(rect.x()),
            y: Math.round(rect.y()),
          };
          onChange?.([update]);
        });
      }

      if (editable) {
        rect.on("dragmove", () => {
          rect.x(Math.round(rect.x() / gridSize) * gridSize);
          rect.y(Math.round(rect.y() / gridSize) * gridSize);
        });
      }

      layer.add(rect);
    });

    // Transformer — attaches to selected rects, exposes resize handles
    const transformer = new Konva.Transformer({
      rotateEnabled: false,
      flipEnabled: false,
      keepRatio: false,
      borderStroke: "#fff",
      anchorStroke: "#fff",
      anchorFill: "#1f2937",
      anchorSize: 8,
      boundBoxFunc: (_oldBox, newBox) => ({
        ...newBox,
        x: Math.round(newBox.x / gridSize) * gridSize,
        y: Math.round(newBox.y / gridSize) * gridSize,
        width: Math.round(newBox.width / gridSize) * gridSize,
        height: Math.round(newBox.height / gridSize) * gridSize,
      }),
    });
    layer.add(transformer);

    const selectedNodes = layer.find("Rect").filter((n) => selectedIds.includes(n.id()));
    transformer.nodes(selectedNodes);

    if (editable) {
      layer.find("Rect").forEach((node) => {
        const rect = node as Konva.Rect;
        rect.on("transformend", () => {
          // Konva applies scale during transform; bake it back into width/height
          const scaleX = rect.scaleX();
          const scaleY = rect.scaleY();
          rect.scaleX(1);
          rect.scaleY(1);
          const update: ZoneUpdate = {
            id: rect.id(),
            x: Math.round(rect.x()),
            y: Math.round(rect.y()),
            width: Math.max(20, Math.round(rect.width() * scaleX)),
            height: Math.max(20, Math.round(rect.height() * scaleY)),
          };
          onChange?.([update]);
        });
      });
    }

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
      const isSelected = selectedIds.includes(rect.id());
      rect.stroke(isSelected ? "#fff" : "transparent");
      rect.strokeWidth(isSelected ? 2 : 0);
    });
    // Re-attach transformer to new selection
    const transformer = layer.findOne<Konva.Transformer>("Transformer");
    if (transformer) {
      const selectedNodes = layer.find("Rect").filter((n) => selectedIds.includes(n.id()));
      transformer.nodes(selectedNodes);
    }
    layer.batchDraw();
  }, [selectedIds]);

  useEffect(() => {
    selectedIdsRef.current = selectedIds;
  }, [selectedIds]);

  return <div ref={containerRef} style={{ width: "100%", height: "100%" }} aria-label="Zone map" />;
}
