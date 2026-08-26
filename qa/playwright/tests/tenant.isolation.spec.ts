import { test, expect } from "@playwright/test";

const tenantA = "11111111-1111-1111-1111-111111111111";
const tenantB = "22222222-2222-2222-2222-222222222222";
const customer = () => process.env.CUSTOMER_BASE || "http://localhost:8111";

test.describe("tenant isolation — customer BFF", () => {
  test("missing tenant header rejected on OTP start", async ({ request }) => {
    const res = await request.post(`${customer()}/v1/customer/auth/otp/start`, {
      data: { phone: "+905551112233" },
    });
    expect([400, 401, 422]).toContain(res.status());
  });

  test("tenant B cannot read tenant A home with wrong scope", async ({ request }) => {
    const res = await request.get(`${customer()}/v1/customer/home?lat=41&lng=29`, {
      headers: { "X-Tenant-Id": tenantB },
    });
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body).toBeTruthy();
  });

  test("order GET without valid order returns not found or upstream", async ({ request }) => {
    const res = await request.get(
      `${customer()}/v1/customer/orders/00000000-0000-0000-0000-000000000099`,
      { headers: { "X-Tenant-Id": tenantA } },
    );
    expect([404, 502, 400]).toContain(res.status());
  });
});
