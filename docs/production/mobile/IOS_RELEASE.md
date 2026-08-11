# iOS Release

## Prerequisites

- Apple Developer Program + App Store Connect apps created
- Distribution certificate + App Store provisioning profiles
- ASC API key in CI (`APP_STORE_CONNECT_API_KEY_*`)

## Build

```bash
cd apps/mobile_customer/ios
flutter build ipa --release --dart-define=ENV=prod
```

## Tracks

1. TestFlight Internal  
2. TestFlight External (Beta App Review if needed)  
3. App Store Review  
4. Phased release (7-day) → 100%

## Review prep

- Privacy nutrition labels (`store/aso/customer/privacy.md`)
- Screenshots / preview (`store/aso/customer/`)
- Demo account for reviewers (OTP_DEV only on review sandbox backend — never prod logging)
- Export compliance / encryption answers documented in ASC

## Crash / perf

- Crashlytics + MetricKit / Xcode Organizer
- Certificate pinning policy per security pack (do not weaken for release)
