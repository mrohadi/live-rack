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
      workbox: {
        cleanupOutdatedCaches: true,
        // The SPA fallback serves index.html for navigations. WITHOUT this
        // denylist the service worker also hijacks full-page navigations to
        // backend/Zitadel routes (e.g. the /oauth/v2/authorize sign-in
        // redirect), serving the cached SPA shell → React Router renders 404.
        // These prefixes must reach the network (Caddy → Zitadel/api) instead.
        // SPA routes like /login, /signup, /callback, /verify-email are NOT
        // listed — they should keep falling back to index.html.
        navigateFallbackDenylist: [
          /^\/oauth\//,
          /^\/oidc\//,
          /^\/openid\//,
          /^\/v2\//,
          /^\/v2beta\//,
          /^\/ui\//,
          /^\/admin\//,
          /^\/system\//,
          /^\/management\//,
          /^\/auth\//,
          /^\/\.well-known\//,
          /^\/device\//,
          /^\/idps\//,
          /^\/grpc\//,
          /^\/debug\//,
          /^\/api\//,
          /^\/ws(?:\/|$)/,
          /^\/healthz$/,
        ],
      },
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
