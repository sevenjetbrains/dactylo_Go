import { defineConfig } from '@playwright/test';

// Tests de bout en bout de l'interface, liaisons Go simulées (e2e/mock-backend.ts).
// Navigateur : Edge installé (Windows) ou Chromium de Playwright en CI.
export default defineConfig({
  testDir: './e2e',
  outputDir: './test-results',
  timeout: 30_000,
  use: {
    baseURL: 'http://localhost:4173',
    channel: process.env.PW_CHANNEL ?? 'msedge',
    viewport: { width: 1100, height: 760 },
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: 'pnpm build && pnpm preview --port 4173 --strictPort',
    url: 'http://localhost:4173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
