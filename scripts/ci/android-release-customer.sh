#!/usr/bin/env bash
# Customer Android App Bundle (release). Uses debug signing when keystore env is unset.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT/apps/mobile_customer"
flutter pub get
flutter build appbundle --release \
  --dart-define=NEXORA_ENV=prod \
  --dart-define=ENV=prod \
  --dart-define=NEXORA_BASE_URL=https://api.nexora.io/v1
ls -la build/app/outputs/bundle/release/*.aab
echo "OK android-aab"
