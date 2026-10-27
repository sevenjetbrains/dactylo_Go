import { defineConfig } from '@playwright/test';

// Essai manuel sur l'application réelle : lancer d'abord
//   $env:DACTYLO_DB = "<base jetable>"; wails dev -browser=false
// puis : pnpm exec playwright test -c playwright.real.config.ts
export default defineConfig({
  testDir: './e2e-real',
  outputDir: './test-results/real',
  workers: 1,
});
