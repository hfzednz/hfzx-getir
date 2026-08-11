# Android Release

## Apps

| App | Module | Application ID |
|-----|--------|----------------|
| Customer | `apps/mobile_customer` | `com.hfzx.nexora.nexora_customer` |
| Courier | `apps/mobile_courier` | see android namespace |
| Warehouse | `apps/mobile_warehouse` | `io.nexora.nexora_warehouse` |

## Versioning

- `pubspec.yaml` `version: X.Y.Z+BUILD` drives `versionName` / `versionCode`.
- BUILD must strictly increase for Play uploads.

## Artifact

```bash
cd apps/mobile_customer
flutter build appbundle --release --dart-define=ENV=prod
# output: build/app/outputs/bundle/release/app-release.aab
```

Fastlane: `apps/mobile_customer/fastlane/Fastfile` lanes `internal`, `closed`, `open`, `production`.

## Play Console tracks

1. Internal testing  
2. Closed testing  
3. Open testing  
4. Production (staged 10% → 50% → 100%)

## Integrity / vitals

- Play Integrity API enabled in Play Console.
- Crashlytics + Play Vitals (ANR, crash-free) monitored post-release.
- Signing: upload key in Play App Signing; CI uses `ANDROID_KEYSTORE_*` secrets (see `android/app/build.gradle.kts` release signingConfigs).

## Workflow

`.github/workflows/cd-mobile.yml` — build AAB, upload to Play track via Fastlane/`google-play` action.
