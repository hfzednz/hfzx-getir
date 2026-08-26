import { test, expect } from "@playwright/test";

const base = () => process.env.CUSTOMER_WEB_BASE || "http://127.0.0.1:3000";

test("customer web login page loads", async ({ page }) => {
  await page.goto(base() + "/login");
  await expect(page.getByRole("heading", { name: /NEXORA/i })).toBeVisible();
  await expect(page.getByLabel(/phone/i)).toBeVisible();
});
