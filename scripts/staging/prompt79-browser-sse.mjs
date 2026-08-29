#!/usr/bin/env node
/**
 * Prompt 79: Chromium EventSource proof with SSE ticket + real warehouse pick.
 * Sequential 390x844 after SSE, then close the browser.
 */
const { chromium } = require("/workspaces/hfzx-getir/qa/playwright/node_modules/@playwright/test");
const { execSync } = require("child_process");
const fs = require("fs");

const ORDER = process.env.PROMPT79_ORDER_ID || fs.readFileSync("/tmp/prompt79-sse-order.txt", "utf8").trim();
const BASE = "http://127.0.0.1:3000";
const PHONE = "+905551112233";
const WH_PHONE = "+905551112234";
const TENANT = "11111111-1111-1111-1111-111111111111";

function otp(phone) {
  const logs = execSync(
    `docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep ${phone} | tail -1`,
    { encoding: "utf8" },
  );
  const m = logs.match(/"code":"(\d+)"/);
  if (!m) throw new Error("no otp for " + phone);
  return m[1];
}

function warehousePick() {
  const start = JSON.parse(
    execSync(
      `curl -sS -H "Content-Type: application/json" -H "X-Tenant-Id: ${TENANT}" -d '{"phone":"${WH_PHONE}","tenantId":"${TENANT}"}' http://127.0.0.1:8081/v1/identity/auth/otp/start`,
      { encoding: "utf8" },
    ),
  );
  const code = otp(WH_PHONE);
  const sess = JSON.parse(
    execSync(
      `curl -sS -H "Content-Type: application/json" -H "X-Tenant-Id: ${TENANT}" -d '{"challengeId":"${start.challengeId}","code":"${code}"}' http://127.0.0.1:8081/v1/identity/auth/otp/verify`,
      { encoding: "utf8" },
    ),
  );
  const tok = sess.accessToken;
  if (!tok) throw new Error("no warehouse token");
  const out = execSync(
    `curl -sS -o /tmp/p79-pick.json -w "%{http_code}" -X POST -H "Content-Type: application/json" -H "X-Tenant-Id: ${TENANT}" -H "Authorization: Bearer ${tok}" -d "{}" http://127.0.0.1:8113/v1/warehouse/tasks/${ORDER}/pick`,
    { encoding: "utf8" },
  );
  console.log("warehouse_pick_http", out.trim());
}

(async () => {
  const browser = await chromium.launch({ headless: true, args: ["--no-sandbox"] });
  const context = await browser.newContext({
    viewport: { width: 390, height: 844 },
    isMobile: true,
    hasTouch: true,
    userAgent:
      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
  });
  const page = await context.newPage();
  page.on("console", (msg) => console.log("BROWSER", msg.type(), msg.text()));

  await page.goto(BASE + "/login", { waitUntil: "networkidle", timeout: 45000 });
  console.log("mobile_overflow_login", await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth));
  await page.getByRole("button", { name: /send otp/i }).waitFor({ timeout: 20000 });
  await page.waitForTimeout(1000);
  await page.getByLabel(/phone/i).fill(PHONE);
  await page.getByRole("button", { name: /send otp/i }).click();
  await page.getByLabel(/otp code/i).waitFor({ timeout: 20000 });
  await page.getByLabel(/otp code/i).fill(otp(PHONE));
  await page.getByRole("button", { name: /verify/i }).click();
  await page.waitForURL(/\/home/, { timeout: 30000 });
  console.log("customer_home_ok");
  console.log("mobile_overflow_home", await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth));

  page.on("response", (res) => {
    const u = res.url();
    if (u.includes("/realtime/") || u.includes("realtime-ticket") || u.includes("/track")) {
      console.log("NET", res.status(), u.replace("http://127.0.0.1:3000", ""));
    }
  });

  await page.goto(BASE + "/orders/" + ORDER + "/track", { waitUntil: "domcontentloaded", timeout: 45000 });
  console.log("track_url", page.url());
  const conn = page.getByTestId("sse-connection");
  await conn.waitFor({ timeout: 20000 });
  await page.waitForFunction(
    () => {
      const el = document.querySelector('[data-testid="sse-connection"]');
      return el && /SSE connected/i.test(el.textContent || "");
    },
    { timeout: 40000 },
  );
  console.log("sse_connection", ((await conn.textContent()) || "").trim());

  warehousePick();
  await page.waitForFunction(
    () => {
      const el = document.querySelector('[data-testid="sse-event-status"]');
      const t = el?.textContent || "";
      return t.includes("Event:") && !t.includes("Event: none");
    },
    { timeout: 25000 },
  );
  console.log("sse_event", ((await page.getByTestId("sse-event-status").textContent()) || "").trim());
  console.log("sse_connection_after", ((await conn.textContent()) || "").trim());

  await page.goto(BASE + "/product/" + "d2ebee03-cb73-4edf-8605-c1f749f5b4fc", { waitUntil: "domcontentloaded", timeout: 30000 });
  console.log("product_url", page.url());
  await page.goto(BASE + "/cart", { waitUntil: "domcontentloaded", timeout: 20000 });
  console.log("cart_ok");
  await page.goto(BASE + "/checkout", { waitUntil: "domcontentloaded", timeout: 20000 });
  console.log("checkout_ok");
  await page.goto(BASE + "/orders", { waitUntil: "domcontentloaded", timeout: 20000 });
  console.log("orders_ok");
  await browser.close();
  console.log("PROMPT79_BROWSER_OK");
})().catch(async (err) => {
  console.error("PROMPT79_BROWSER_FAIL", err);
  process.exit(1);
});
