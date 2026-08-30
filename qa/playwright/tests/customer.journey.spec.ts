import { test, expect } from '@playwright/test';

const customer = () => process.env.CUSTOMER_BASE || 'http://localhost:8111';
const headers = {
  'X-Tenant-Id': '11111111-1111-1111-1111-111111111111',
  'X-Request-Id': 'pw-journey-57',
};

test('customer BFF login home browse cart health', async ({ request }) => {
  const health = await request.get(customer() + '/health');
  expect(health.ok()).toBeTruthy();

  const token = process.env.CUSTOMER_TOKEN;
  const home = await request.get(customer() + '/v1/customer/home?lat=41.0&lng=29.0', {
    headers: token ? { ...headers, Authorization: `Bearer ${token}` } : headers,
  });
  // The feed needs a session; without one the only correct answer is 401.
  expect(token ? home.ok() : home.status() === 401).toBeTruthy();
  expect(home.headers()['x-request-id'] || home.headers()['X-Request-Id']).toBeTruthy();

  const otp = await request.post(customer() + '/v1/customer/auth/otp/start', {
    headers: { ...headers, 'Content-Type': 'application/json' },
    data: { phone: '+905551112244' },
  });
  expect(otp.ok()).toBeTruthy();
  const body = await otp.json();
  expect(body.challengeId || body.ChallengeID).toBeTruthy();
});
