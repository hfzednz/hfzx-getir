import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 60_000,
  retries: 1,
  use: {
    baseURL: process.env.ADMIN_BASE || 'http://localhost:8114',
    trace: 'on-first-retry',
  },
  reporter: [['list'], ['junit', { outputFile: '../reports/playwright-junit.xml' }], ['html', { open: 'never' }]],
});
