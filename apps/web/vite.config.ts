import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: ["favicon.ico"],
      manifest: {
        name: "live-rack Scanner",
        short_name: "Scanner",
        start_url: "/scanner",
        display: "standalone",
        background_color: "#0f172a",
        theme_color: "#0f172a",
        icons: [
          { src: "/pwa-192.png", sizes: "192x192", type: "image/png" },
          { src: "/pwa-512.png", sizes: "512x512", type: "image/png" },
        ],
      },
    }),
  ],
  server: { port: 5173, proxy: { "/api": "http://localhost:8080" } },
  resolve: {
    alias: {
      konva: "konva/lib/index.js",
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
