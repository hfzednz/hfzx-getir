#!/usr/bin/env bash
# Typecheck, unit-test and build all NEXORA web applications (does not deploy).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export NEXT_TELEMETRY_DISABLED=1

WEB_APPS=(
  customer-web
  courier-web
  warehouse-web
  supplier-web
  finance-web
  support-web
  operations-web
  admin_web
  super_admin_web
)

echo "==> @nexora/web-core"
cd "$ROOT/packages/web/core"
npm install --no-fund --no-audit

echo "==> @nexora/ui"
cd "$ROOT/packages/web/ui"
npm install --no-fund --no-audit

for app in "${WEB_APPS[@]}"; do
  dir="$ROOT/apps/$app"
  if [[ ! -f "$dir/package.json" ]]; then
    echo "SKIP $app (no package.json)"
    continue
  fi
  echo "==> $app"
  cd "$dir"
  if [[ -f package-lock.json ]]; then
    npm ci --no-fund --no-audit
  else
    npm install --no-fund --no-audit
  fi
  has_script() {
    node -e "process.exit(require('./package.json').scripts?.['$1'] ? 0 : 1)"
  }
  if has_script typecheck; then
    npm run typecheck
  else
    npx --yes tsc --noEmit -p tsconfig.json
  fi
  if has_script test; then
    npm test
  fi
  npm run build
done

echo "WEB_STATIC_PASS"
