#!/usr/bin/env bash
# Web E2E: disposable phone-test stack + customer-web + Playwright.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export NEXT_TELEMETRY_DISABLED=1
export CUSTOMER_BASE="http://127.0.0.1:8111"
export CUSTOMER_WEB_BASE="http://127.0.0.1:3000"

echo "==> disposable checkout stack"
bash "$ROOT/scripts/staging/deploy-phone-test.sh"

echo "==> customer-web"
cd "$ROOT/packages/web/core" && npm install --no-fund --no-audit
cd "$ROOT/packages/web/ui" && npm install --no-fund --no-audit
cd "$ROOT/apps/customer-web"
npm ci --no-fund --no-audit 2>/dev/null || npm install --no-fund --no-audit
npm run build
npm run start -- -p 3000 &
web_pid=$!
cleanup() { kill "$web_pid" 2>/dev/null || true; }
trap cleanup EXIT

for _ in $(seq 1 40); do
  curl -fsS "$CUSTOMER_WEB_BASE/login" >/dev/null 2>&1 && break
  sleep 2
done
curl -fsS "$CUSTOMER_WEB_BASE/login" | grep -qi nexora

echo "==> playwright"
cd "$ROOT/qa/playwright"
npm install --no-fund --no-audit
npx playwright install --with-deps chromium
npx playwright test tests/customer.web.login.spec.ts tests/tenant.isolation.spec.ts tests/multi-role.journey.spec.ts --project=api --reporter=list
npx playwright test tests/customer.web.login.spec.ts --project=ui-customer --reporter=list

echo "WEB_E2E_PASS"
