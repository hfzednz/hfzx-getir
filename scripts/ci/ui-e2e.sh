#!/usr/bin/env bash
# Admin web UI + accessibility Playwright (HTML surface). Customer Flutter checkout is not in this job.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export NEXT_TELEMETRY_DISABLED=1
PORT=3100
export ADMIN_WEB_BASE="http://127.0.0.1:${PORT}"

cd "$ROOT/packages/web/ui"
npm install --no-fund --no-audit
cd "$ROOT/apps/admin_web"
npm ci --no-fund --no-audit
npm run build
npm run start -- -p "$PORT" &
pid=$!
cleanup() { kill "$pid" >/dev/null 2>&1 || true; }
trap cleanup EXIT

ok=0
for _ in $(seq 1 60); do
  if curl -fsS --max-time 2 "$ADMIN_WEB_BASE/login" >/dev/null 2>&1; then
    ok=1
    break
  fi
  sleep 2
done
if [[ "$ok" -ne 1 ]]; then
  echo "FAIL admin_web did not start on :$PORT"
  exit 1
fi

cd "$ROOT/qa/playwright"
npm install --no-fund --no-audit
npx playwright install --with-deps chromium
npx playwright test --project=ui --reporter=list
echo "UI_E2E_PASS"
