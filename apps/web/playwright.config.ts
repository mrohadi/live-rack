import { defineConfig, devices } from "@playwright/test";
import { parse } from "dotenv";
import { readFileSync } from "fs";
import { resolve } from "path";

try {
  const parsed = parse(readFileSync(resolve(process.cwd(), ".env.local")));
  Object.assign(process.env, parsed);
} catch {}
try {
  const parsed = parse(readFileSync(resolve(process.cwd(), ".env")));
  Object.assign(process.env, parsed);
} catch {}

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: "list",
  globalSetup: "./e2e/global-setup.ts",
  use: {
    baseURL: "http://localhost:5173",
    trace: "on-first-retry",
    // This passes the slowMo option straight to the browser launch context
    launchOptions: {
      slowMo: 500, // Delays operations by 500ms
    },
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  // Start Vite dev server before running tests
  webServer: {
    command: "pnpm dev",
    url: "http://localhost:5173",
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
