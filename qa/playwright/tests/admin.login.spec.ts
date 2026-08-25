import { test, expect } from '@playwright/test';

const web = process.env.ADMIN_WEB_BASE;

test.describe('admin web UI', () => {
  test.skip(!web, 'ADMIN_WEB_BASE not set');

  test.use({ baseURL: web, extraHTTPHeaders: {} });

  test('invalid login stays on login with empty password', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: /NEXORA Admin/i })).toBeVisible();
    await page.locator('input[type="password"]').fill('');
    await page.getByRole('button', { name: /Sign in/i }).click();
    await expect(page).toHaveURL(/\/login/);
  });

  test('login home orders inventory finance', async ({ page }) => {
    await page.goto('/login');
    await page.locator('input[type="email"]').fill('ops@nexora.local');
    await page.locator('input[type="password"]').fill('demo');
    await page.getByRole('button', { name: /Sign in/i }).click();
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 15_000 });
    await page.goto('/orders');
    await expect(page).toHaveURL(/\/orders/);
    await page.goto('/inventory');
    await expect(page).toHaveURL(/\/inventory/);
    await page.goto('/finance');
    await expect(page).toHaveURL(/\/finance/);
  });

  test('keyboard focus order on login form', async ({ page }) => {
    await page.goto('/login');
    await page.keyboard.press('Tab');
    const email = page.locator('input[type="email"]');
    await email.focus();
    await expect(email).toBeFocused();
    await page.keyboard.press('Tab');
    await expect(page.locator('input[type="password"]')).toBeFocused();
  });
});
