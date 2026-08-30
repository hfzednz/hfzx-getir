import { test, expect } from "@playwright/test";

const tenantA = "11111111-1111-1111-1111-111111111111";
const tenantB = "22222222-2222-2222-2222-222222222222";
const customer = () => process.env.CUSTOMER_BASE || "http://localhost:8111";
const customerToken = () => process.env.CUSTOMER_TOKEN || "";

function authHeaders(tenant: string) {
  const headers: Record<string, string> = { "X-Tenant-Id": tenant };
  const token = customerToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

test.describe("tenant isolation — customer BFF", () => {
  test("missing tenant header rejected on OTP start", async ({ request }) => {
    const res = await request.post(`${customer()}/v1/customer/auth/otp/start`, {
      data: { phone: "+905551112233" },
    });
    expect([400, 401, 422]).toContain(res.status());
  });

  test("storefront feed requires a session", async ({ request }) => {
    const res = await request.get(`${customer()}/v1/customer/home?lat=41&lng=29`, {
      headers: { "X-Tenant-Id": tenantA },
    });
    expect(res.status()).toBe(401);
  });

  test("a tenant A session cannot read the feed as tenant B", async ({ request }) => {
    test.skip(!customerToken(), "CUSTOMER_TOKEN not set");
    const res = await request.get(`${customer()}/v1/customer/home?lat=41&lng=29`, {
      headers: authHeaders(tenantB),
    });
    expect([401, 403, 404]).toContain(res.status());
  });

  test("unknown order id is not readable", async ({ request }) => {
    const res = await request.get(
      `${customer()}/v1/customer/orders/00000000-0000-0000-0000-000000000099`,
      { headers: authHeaders(tenantA) },
    );
    expect([401, 404, 502, 400]).toContain(res.status());
  });
});
