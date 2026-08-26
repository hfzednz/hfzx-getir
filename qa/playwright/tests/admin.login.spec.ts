import { test, expect } from '@playwright/test';

const web = process.env.ADMIN_WEB_BASE;

test.describe('admin web UI', () => {
  test.skip(!web, 'ADMIN_WEB_BASE not set');

  test.use({ baseURL: web, extraHTTPHeaders: {} });

  test('login page shows OTP phone field', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: /NEXORA Admin/i })).toBeVisible();
    await expect(page.getByLabel(/phone/i)).toBeVisible();
  });

  test('keyboard focus on phone field', async ({ page }) => {
    await page.goto('/login');
    const phone = page.getByLabel(/phone/i);
    await phone.focus();
    await expect(phone).toBeFocused();
  });
});
