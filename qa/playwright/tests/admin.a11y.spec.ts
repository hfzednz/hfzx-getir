import { test, expect } from '@playwright/test';

const web = process.env.ADMIN_WEB_BASE;

test.describe('admin accessibility', () => {
  test.skip(!web, 'ADMIN_WEB_BASE not set');
  test.use({ baseURL: web });

  test('OTP login form has labels heading and named submit', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: /NEXORA Admin/i })).toBeVisible();
    await expect(page.getByText('Phone', { exact: true })).toBeVisible();
    const phone = page.locator('input[type="tel"]');
    await expect(phone).toHaveAttribute('autocomplete', 'tel');
    const submit = page.getByRole('button', { name: /Send OTP/i });
    await expect(submit).toBeVisible();
    const box = await submit.boundingBox();
    expect(box).toBeTruthy();
    if (box) {
      expect(box.height).toBeGreaterThanOrEqual(24);
      expect(box.width).toBeGreaterThanOrEqual(44);
    }
  });
});
