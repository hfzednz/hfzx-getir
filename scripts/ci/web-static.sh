#!/usr/bin/env bash
# Build + lint all NEXORA web applications (does not deploy).
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
  npm run build
done

echo "WEB_STATIC_PASS"
