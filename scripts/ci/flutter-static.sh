#!/usr/bin/env bash
# Flutter static gates: pub get, analyze (non-fatal infos), unit/widget tests.
# Does not require an emulator. Live BFF tests skip unless CUSTOMER_BASE is set.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export PUB_CACHE="${PUB_CACHE:-$HOME/.pub-cache}"

run_app() {
  local dir="$1"
  echo "==> flutter $dir"
  (
    cd "$ROOT/$dir"
    flutter pub get
    flutter analyze --no-fatal-infos --no-fatal-warnings
    flutter test --reporter compact
  )
}

run_app apps/mobile_customer
run_app apps/mobile_courier
run_app apps/mobile_warehouse
echo "OK flutter-static"
