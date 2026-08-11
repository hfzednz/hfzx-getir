import { test, expect } from '@playwright/test';

test('admin BFF health', async ({ request }) => {
  const res = await request.get('/health');
  expect(res.ok()).toBeTruthy();
});
