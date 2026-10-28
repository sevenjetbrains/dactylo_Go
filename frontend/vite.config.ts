// Configuration Vite : Svelte, Tailwind, alias d'import et tests Vitest.

import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { svelteTesting } from '@testing-library/svelte/vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [tailwindcss(), svelte(), svelteTesting()],
  resolve: {
    alias: {
      $app: fileURLToPath(new URL('./src/app', import.meta.url)),
      $modules: fileURLToPath(new URL('./src/modules', import.meta.url)),
      $shared: fileURLToPath(new URL('./src/shared', import.meta.url)),
      $i18n: fileURLToPath(new URL('./src/i18n', import.meta.url)),
      $wails: fileURLToPath(new URL('./wailsjs', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.ts'],
  },
});
