#!/usr/bin/env node
/**
 * Chromium EventSource proof against customer-web tracking page.
 * Requires PROMPT78_ORDER_ID and a warehouse pick after the page is open.
 */
const { chromium } = require("/workspaces/hfzx-getir/qa/playwright/node_modules/@playwright/test");
const { execSync } = require("child_process");
const fs = require("fs");

const ORDER = process.env.PROMPT78_ORDER_ID || fs.readFileSync("/tmp/prompt78-order-id.txt", "utf8").trim();
const BASE = "http://127.0.0.1:3000";
const PHONE = "+905551112233";
const TENANT = "11111111-1111-1111-1111-111111111111";

function otp() {
  const logs = execSync(
    "docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep +905551112233 | tail -1",
    { encoding: "utf8" },
  );
  const m = logs.match(/"code":"(\d+)"/);
  if (!m) throw new Error("no otp");
  return m[1];
}

(async () => {
  const mobile = process.env.PROMPT78_MOBILE === "1";
  const browser = await chromium.launch({ headless: true, args: ["--no-sandbox"] });
  const context = await browser.newContext(
    mobile
      ? { viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true, userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1" }
      : {},
  );
  const page = await context.newPage();
  page.on("console", (msg) => console.log("BROWSER", msg.type(), msg.text()));

  await page.goto(BASE + "/login", { waitUntil: "domcontentloaded", timeout: 45000 });
  if (mobile) {
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    console.log("mobile_overflow_login", overflow);
  }
  await page.getByLabel(/phone/i).fill(PHONE);
  await page.getByRole("button", { name: /send otp/i }).click();
  await page.getByLabel(/otp code/i).waitFor({ timeout: 20000 });
  await page.getByLabel(/otp code/i).fill(otp());
  await page.getByRole("button", { name: /verify/i }).click();
  await page.waitForURL(/\/home/, { timeout: 30000 });
  console.log("customer_home_ok");
  if (mobile) {
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    console.log("mobile_overflow_home", overflow);
    const phone = await page.getByLabel(/search products/i).boundingBox();
    console.log("search_touch_h", phone && phone.height);
  }

  await page.goto(BASE + "/orders/" + ORDER + "/track", { waitUntil: "domcontentloaded", timeout: 45000 });
  const conn = page.getByTestId("sse-connection");
  await conn.waitFor({ timeout: 20000 });
  // Wait until EventSource onopen flips the label.
  await page.waitForFunction(
    () => {
      const el = document.querySelector('[data-testid="sse-connection"]');
      return el && /SSE connected/i.test(el.textContent || "");
    },
    { timeout: 25000 },
  );
  console.log("sse_connection", (await conn.textContent()) || "");

  const before = (await page.getByTestId("sse-event-status").textContent()) || "";
  console.log("sse_event_before", before.trim());

  // Legitimate warehouse pick/pack on a NEW order is done by the python driver.
  // This script only proves the customer tracking page EventSource is open + can receive.
  // Trigger a real publish by posting warehouse pick if still warehouse_assigned, else a second event via courier if supported.
  const { execFileSync } = require("child_process");
  try {
    execFileSync("curl", [
      "-sS", "-X", "POST",
      "-H", "Content-Type: application/json",
      "-H", "X-Tenant-Id: " + TENANT,
      "http://127.0.0.1:8113/v1/warehouse/tasks/" + ORDER + "/pick",
      "-d", "{}",
    ], { encoding: "utf8", timeout: 15000 });
  } catch (e) {
    console.log("warehouse_pick_note", String(e.stdout || e.message || e).slice(0, 200));
  }

  await page.waitForFunction(
    () => {
      const el = document.querySelector('[data-testid="sse-event-status"]');
      return el && !/Event: none/i.test(el.textContent || "");
    },
    { timeout: 20000 },
  ).then(
    async () => {
      console.log("sse_event_after", ((await page.getByTestId("sse-event-status").textContent()) || "").trim());
      console.log("SSE_BROWSER_EVENT PASS");
    },
    async () => {
      console.log("sse_event_after", ((await page.getByTestId("sse-event-status").textContent()) || "").trim());
      console.log("SSE_BROWSER_EVENT FAIL");
    },
  );

  await browser.close();
})().catch((err) => {
  console.error(err);
  process.exit(1);
});
