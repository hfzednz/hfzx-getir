import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 60_000,
  retries: 1,
  use: {
    trace: 'on-first-retry',
  },
  reporter: [['list'], ['junit', { outputFile: '../reports/playwright-junit.xml' }], ['html', { open: 'never' }]],
  projects: [
    {
      name: 'api',
      testMatch:
        /customer\.(health|negative|journey)\.spec\.ts|admin\.health\.spec\.ts|tenant\.isolation\.spec\.ts/,
      use: {
        baseURL: process.env.ADMIN_BASE || 'http://localhost:8114',
      },
    },
    {
      name: 'ui',
      testMatch: /admin\.(login|a11y)\.spec\.ts|customer\.web\.login\.spec\.ts/,
      use: {
        baseURL:
          process.env.CUSTOMER_WEB_BASE ||
          process.env.ADMIN_WEB_BASE ||
          'http://127.0.0.1:3100',
        browserName: 'chromium',
      },
    },
  ],
});
