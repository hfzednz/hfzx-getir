import { test, expect } from '@playwright/test';

const web = process.env.ADMIN_WEB_BASE;

test.describe('admin accessibility', () => {
  test.skip(!web, 'ADMIN_WEB_BASE not set');
  test.use({ baseURL: web });

  test('login form has labels heading and named submit', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: /NEXORA Admin/i })).toBeVisible();
    await expect(page.getByText('Email', { exact: true })).toBeVisible();
    await expect(page.getByText('Password', { exact: true })).toBeVisible();
    const email = page.locator('input[type="email"]');
    await expect(email).toHaveAttribute('autocomplete', 'username');
    const password = page.locator('input[type="password"]');
    await expect(password).toHaveAttribute('autocomplete', 'current-password');
    const submit = page.getByRole('button', { name: /Sign in/i });
    await expect(submit).toBeVisible();
    const box = await submit.boundingBox();
    expect(box).toBeTruthy();
    if (box) {
      expect(box.height).toBeGreaterThanOrEqual(24);
      expect(box.width).toBeGreaterThanOrEqual(44);
    }
  });
});
