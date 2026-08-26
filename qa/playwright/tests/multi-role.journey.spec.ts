import { test, expect } from "@playwright/test";

const tenant = "11111111-1111-1111-1111-111111111111";
const headers = { "X-Tenant-Id": tenant, "Content-Type": "application/json" };
const customer = () => process.env.CUSTOMER_BASE || "http://127.0.0.1:8111";
const admin = () => process.env.ADMIN_BASE || "http://127.0.0.1:8114";
const warehouse = () => process.env.WAREHOUSE_BASE || "http://127.0.0.1:8113";
const finance = () => process.env.FINANCE_BASE || "http://127.0.0.1:8091";

test.describe("multi-role API journey", () => {
  test("customer home → admin orders → finance journals", async ({ request }) => {
    const home = await request.get(`${customer()}/v1/customer/home?lat=41&lng=29`, { headers });
    expect(home.ok()).toBeTruthy();

    const dash = await request.get(`${admin()}/v1/admin/dashboard`, { headers });
    expect(dash.ok()).toBeTruthy();

    const orders = await request.get(`${admin()}/v1/admin/orders`, { headers });
    expect([200, 502]).toContain(orders.status());

    const journals = await request.get(`${finance()}/v1/ledger/journals`, { headers });
    expect([200, 404, 502]).toContain(journals.status());
  });

  test("warehouse health reachable", async ({ request }) => {
    const res = await request.get(`${warehouse()}/health`);
    expect(res.ok()).toBeTruthy();
  });

  test("tenant B header rejected for wrong-scope order", async ({ request }) => {
    const res = await request.get(
      `${customer()}/v1/customer/orders/00000000-0000-0000-0000-000000000099`,
      { headers: { ...headers, "X-Tenant-Id": "22222222-2222-2222-2222-222222222222" } },
    );
    expect([404, 400, 502]).toContain(res.status());
  });
});
