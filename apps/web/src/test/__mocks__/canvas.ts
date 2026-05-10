// Stub for Konva's node build: `require('canvas')` in konva/lib/index-node.js
class CanvasStub {
  width = 800;
  height = 600;
  getContext() {
    return {
      clearRect() {},
      fillRect() {},
      strokeRect() {},
      beginPath() {},
      closePath() {},
      moveTo() {},
      lineTo() {},
      stroke() {},
      fill() {},
      arc() {},
      rect() {},
      clip() {},
      save() {},
      restore() {},
      translate() {},
      scale() {},
      rotate() {},
      setTransform() {},
      transform() {},
      drawImage() {},
      measureText() {
        return { width: 0 };
      },
      fillText() {},
      strokeText() {},
      createLinearGradient() {
        return { addColorStop() {} };
      },
      createPattern() {
        return {};
      },
      putImageData() {},
      getImageData() {
        return { data: new Uint8ClampedArray() };
      },
    };
  }
}

class ImageStub {}
class DOMMatrixStub {
  constructor() {}
  a = 1;
  b = 0;
  c = 0;
  d = 1;
  e = 0;
  f = 0;
}

export function createCanvas(width: number, height: number) {
  const c = new CanvasStub();
  c.width = width;
  c.height = height;
  return c;
}

export const Canvas = CanvasStub;
export const Image = ImageStub;
export const DOMMatrix = DOMMatrixStub;

export default { createCanvas, Canvas: CanvasStub, Image: ImageStub, DOMMatrix: DOMMatrixStub };
