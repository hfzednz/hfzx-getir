import { test, expect } from '@playwright/test';

test('customer BFF health', async ({ request }) => {
  const base = process.env.CUSTOMER_BASE || 'http://localhost:8111';
  const res = await request.get(base + '/health');
  expect(res.ok()).toBeTruthy();
});
