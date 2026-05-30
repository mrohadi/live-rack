import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    globals: true,
    exclude: ["node_modules", "dist", "e2e/**"], // ← top-level, NOT inside coverage
    alias: {
      konva: "konva/lib/index.js",
    },
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov", "html"],
      reportsDirectory: "./coverage",
      include: ["src/features/**", "src/lib/**", "src/components/**"],
      exclude: [
        "src/test/**",
        "src/**/*.test.{ts,tsx}",
        "src/main.tsx",
        "src/routes.tsx",
        "node_modules",
        "dist",
        "e2e/**",
      ],
      thresholds: { lines: 0, functions: 0, branches: 0, statements: 0 },
    },
  },
});
