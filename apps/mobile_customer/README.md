# NEXORA Customer Mobile App

Enterprise quick-commerce customer application (Flutter). Governed by:

- [`docs/constitution/MASTER_BLUEPRINT.md`](../../docs/constitution/MASTER_BLUEPRINT.md)
- [`docs/design-system/00-INDEX.md`](../../docs/design-system/00-INDEX.md)
- [`ARCHITECTURE.md`](ARCHITECTURE.md) (this app)

## Stack

Flutter · Riverpod · GoRouter · Dio · Drift · Hive · Secure Storage · FCM · Maps · WebSocket · speech_to_text · mobile_scanner · app_links

## Modules

Auth, Onboarding, Splash, Home, Search (voice/barcode/image), Categories, Product, Cart (offline-first), Checkout, Orders, Live Tracking, Favorites, Wallet, Coupons, Loyalty, Notifications, Profile, Addresses, Reviews, Support, Settings, Referral, Help, About, Legal, City

## Setup

```bash
cd apps/mobile_customer
cp .env.example .env   # optional; dart-define also supported
flutter pub get
dart run build_runner build --delete-conflicting-outputs
flutter gen-l10n
```

## Run

```bash
flutter run

flutter run \
  --dart-define=NEXORA_BASE_URL=https://api.dev.nexora.local/v1 \
  --dart-define=NEXORA_WS_URL=wss://realtime.dev.nexora.local/v1 \
  --dart-define=NEXORA_ENV=dev
```

| Define | Example |
|--------|---------|
| `NEXORA_BASE_URL` | `https://api.nexora.io/v1` |
| `NEXORA_WS_URL` | `wss://realtime.nexora.io/v1` |
| `NEXORA_ENV` | `prod` |
| `NEXORA_DEFAULT_LANGUAGE` | `tr` |

## Fonts

Drop licensed **Satoshi** / **Geist** / **Geist Mono** into `assets/fonts/` and register in `pubspec.yaml`. Until then, `nexora_design` system fallbacks apply.

## Firebase

Optional in bootstrap — missing platform config skips push/crashlytics gracefully.

## Tests

```bash
flutter test
flutter test integration_test/app_boot_test.dart
```

## Deep links

Handled by `DeepLinkHandler` (`app_links`):

- `/p/:productId`, `/orders/:orderId`, `/orders/:orderId/track`
- `/c/:slug`, `/promo/:code`, `/cart`, `/search`
- `nexora://product/{id}`
