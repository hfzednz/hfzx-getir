import { test, expect } from '@playwright/test';

const customer = () => process.env.CUSTOMER_BASE || 'http://localhost:8111';
const tenant = { 'X-Tenant-Id': '11111111-1111-1111-1111-111111111111' };

test('invalid login otp remains unauthenticated', async ({ request }) => {
  const start = await request.post(customer() + '/v1/customer/auth/otp/start', {
    headers: { ...tenant, 'Content-Type': 'application/json' },
    data: { phone: '+905551112233' },
  });
  expect(start.ok()).toBeTruthy();
  const { challengeId } = await start.json();
  const bad = await request.post(customer() + '/v1/customer/auth/otp/verify', {
    headers: { ...tenant, 'Content-Type': 'application/json' },
    data: { challengeId, code: '000000' },
  });
  expect(bad.status()).toBeGreaterThanOrEqual(400);
  const health = await request.get(customer() + '/health');
  expect(health.ok()).toBeTruthy();
});

test('otp start without phone is rejected', async ({ request }) => {
  const res = await request.post(customer() + '/v1/customer/auth/otp/start', {
    headers: { ...tenant, 'Content-Type': 'application/json' },
    data: { phone: '' },
  });
  expect(res.status()).toBe(400);
});

test('otp start without tenant is rejected', async ({ request }) => {
  const res = await request.post(customer() + '/v1/customer/auth/otp/start', {
    headers: { 'Content-Type': 'application/json' },
    data: { phone: '+905551112233' },
  });
  expect(res.status()).toBe(400);
});

test('empty cart add is rejected and health stays ok', async ({ request }) => {
  const res = await request.post(customer() + '/v1/customer/cart/items', {
    headers: { ...tenant, 'Content-Type': 'application/json' },
    data: { cartId: '', sku: 'nope', qty: 1, unitMinor: 100 },
  });
  expect(res.status()).toBeGreaterThanOrEqual(400);
  const health = await request.get(customer() + '/health');
  expect(health.ok()).toBeTruthy();
});
