#!/usr/bin/env bash
# Customer Android staging APK + AAB (debug-signed when keystore env is unset).
# Requires NEXORA_STAGING_BASE_URL — the public HTTPS customer BFF base (e.g. https://api-staging.example.com/v1).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT/apps/mobile_customer"

BASE_URL="${NEXORA_STAGING_BASE_URL:-}"
if [[ -z "$BASE_URL" ]]; then
  echo "FAIL: set NEXORA_STAGING_BASE_URL to the public staging customer BFF (https://host/v1)" >&2
  exit 1
fi

WS_URL="${NEXORA_STAGING_WS_URL:-}"
if [[ -z "$WS_URL" ]]; then
  # Derive wss from https host when not supplied.
  host="${BASE_URL#https://}"
  host="${host#http://}"
  host="${host%%/*}"
  WS_URL="wss://${host}/v1"
fi

echo "==> staging customer Android (APK + AAB)"
echo "    NEXORA_BASE_URL=$BASE_URL"
echo "    NEXORA_WS_URL=$WS_URL"

flutter pub get
flutter build apk --release \
  --dart-define=NEXORA_ENV=staging \
  --dart-define=ENV=staging \
  --dart-define=NEXORA_BASE_URL="$BASE_URL" \
  --dart-define=NEXORA_WS_URL="$WS_URL"
flutter build appbundle --release \
  --dart-define=NEXORA_ENV=staging \
  --dart-define=ENV=staging \
  --dart-define=NEXORA_BASE_URL="$BASE_URL" \
  --dart-define=NEXORA_WS_URL="$WS_URL"

ls -la build/app/outputs/flutter-apk/*.apk
ls -la build/app/outputs/bundle/release/*.aab
echo "OK android-staging-apk-aab"
