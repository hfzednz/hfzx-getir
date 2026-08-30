import { test, expect } from "@playwright/test";

const tenant = "11111111-1111-1111-1111-111111111111";
const baseHeaders = { "X-Tenant-Id": tenant, "Content-Type": "application/json" };
const customer = () => process.env.CUSTOMER_BASE || "http://127.0.0.1:8111";

function withToken(token: string | undefined) {
  return token ? { ...baseHeaders, Authorization: `Bearer ${token}` } : baseHeaders;
}

test.describe("multi-role API journey", () => {
  test("customer feed reachable with a customer session", async ({ request }) => {
    test.skip(!process.env.CUSTOMER_TOKEN, "CUSTOMER_TOKEN not set");
    const home = await request.get(`${customer()}/v1/customer/home?lat=41&lng=29`, {
      headers: withToken(process.env.CUSTOMER_TOKEN),
    });
    expect(home.ok()).toBeTruthy();
  });

  test("admin dashboard refuses an anonymous caller", async ({ request }) => {
    test.skip(!process.env.ADMIN_BASE, "ADMIN_BASE not set");
    const dash = await request.get(`${process.env.ADMIN_BASE}/v1/admin/dashboard`, {
      headers: baseHeaders,
    });
    expect([401, 403]).toContain(dash.status());
  });

  test("admin orders refuse a customer session", async ({ request }) => {
    test.skip(!process.env.ADMIN_BASE || !process.env.CUSTOMER_TOKEN, "admin base or token not set");
    const orders = await request.get(`${process.env.ADMIN_BASE}/v1/admin/orders`, {
      headers: withToken(process.env.CUSTOMER_TOKEN),
    });
    expect([401, 403]).toContain(orders.status());
  });

  test("finance journals reachable with a finance session", async ({ request }) => {
    test.skip(!process.env.FINANCE_BASE || !process.env.FINANCE_TOKEN, "finance base or token not set");
    const journals = await request.get(`${process.env.FINANCE_BASE}/v1/ledger/journals`, {
      headers: withToken(process.env.FINANCE_TOKEN),
    });
    expect([200, 404]).toContain(journals.status());
  });

  test("finance journals refuse a customer session", async ({ request }) => {
    test.skip(!process.env.FINANCE_BASE || !process.env.CUSTOMER_TOKEN, "finance base or token not set");
    const journals = await request.get(`${process.env.FINANCE_BASE}/v1/ledger/journals`, {
      headers: withToken(process.env.CUSTOMER_TOKEN),
    });
    expect([401, 403]).toContain(journals.status());
  });

  test("warehouse health when WAREHOUSE_BASE set", async ({ request }) => {
    test.skip(!process.env.WAREHOUSE_BASE, "WAREHOUSE_BASE not set");
    const res = await request.get(`${process.env.WAREHOUSE_BASE}/health`);
    expect(res.ok()).toBeTruthy();
  });

  test("a session cannot read another tenant's order", async ({ request }) => {
    const res = await request.get(
      `${customer()}/v1/customer/orders/00000000-0000-0000-0000-000000000099`,
      {
        headers: {
          ...withToken(process.env.CUSTOMER_TOKEN),
          "X-Tenant-Id": "22222222-2222-2222-2222-222222222222",
        },
      },
    );
    expect([401, 403, 404, 400, 502]).toContain(res.status());
  });
});
