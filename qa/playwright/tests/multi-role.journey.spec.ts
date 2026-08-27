import { test, expect } from "@playwright/test";

const tenant = "11111111-1111-1111-1111-111111111111";
const headers = { "X-Tenant-Id": tenant, "Content-Type": "application/json" };
const customer = () => process.env.CUSTOMER_BASE || "http://127.0.0.1:8111";

test.describe("multi-role API journey", () => {
  test("customer home reachable", async ({ request }) => {
    const home = await request.get(`${customer()}/v1/customer/home?lat=41&lng=29`, { headers });
    expect(home.ok()).toBeTruthy();
  });

  test("admin dashboard when ADMIN_BASE set", async ({ request }) => {
    test.skip(!process.env.ADMIN_BASE, "ADMIN_BASE not set");
    const dash = await request.get(`${process.env.ADMIN_BASE}/v1/admin/dashboard`, { headers });
    expect(dash.ok()).toBeTruthy();
  });

  test("admin orders when ADMIN_BASE set", async ({ request }) => {
    test.skip(!process.env.ADMIN_BASE, "ADMIN_BASE not set");
    const orders = await request.get(`${process.env.ADMIN_BASE}/v1/admin/orders`, { headers });
    expect([200, 502]).toContain(orders.status());
  });

  test("finance journals when FINANCE_BASE set", async ({ request }) => {
    test.skip(!process.env.FINANCE_BASE, "FINANCE_BASE not set");
    const journals = await request.get(`${process.env.FINANCE_BASE}/v1/ledger/journals`, { headers });
    expect([200, 404, 502]).toContain(journals.status());
  });

  test("warehouse health when WAREHOUSE_BASE set", async ({ request }) => {
    test.skip(!process.env.WAREHOUSE_BASE, "WAREHOUSE_BASE not set");
    const res = await request.get(`${process.env.WAREHOUSE_BASE}/health`);
    expect(res.ok()).toBeTruthy();
  });

  test("tenant B order scope", async ({ request }) => {
    const res = await request.get(
      `${customer()}/v1/customer/orders/00000000-0000-0000-0000-000000000099`,
      { headers: { ...headers, "X-Tenant-Id": "22222222-2222-2222-2222-222222222222" } },
    );
    expect([404, 400, 502]).toContain(res.status());
  });
});
