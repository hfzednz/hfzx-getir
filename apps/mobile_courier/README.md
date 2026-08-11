# NEXORA Courier (`nexora_courier`)

Enterprise courier client for the NEXORA delivery platform. Compact, task-first UI on Flutter + Riverpod + GoRouter, sharing `nexora_core` and `nexora_design`.

## Stack

| Concern | Choice |
|---------|--------|
| State | Riverpod |
| Routing | GoRouter |
| HTTP | Dio via `ApiClient` (`apiClientProvider`) |
| Maps | `google_maps_flutter` + `url_launcher` external nav |
| Scan | `mobile_scanner` |
| POD photo | `image_picker` |
| Battery | `battery_plus` |
| Realtime | `RealtimeClient` (`offer.*` invalidates offers) |
| BFF | `bff-courier` at `/v1` |

## Feature modules

Each feature follows `domain/` · `data/` · `presentation/`:

| Module | Screens / notes | Primary APIs |
|--------|-----------------|--------------|
| **status** | `DutyController`, `StatusControlWidget`, `StatusScreen` | `POST /courier/duty/status` |
| **home** | `DashboardScreen` / `HomeScreen` | `GET /courier/dashboard` |
| **offers** | Accept/reject with idempotency + realtime | `GET/POST /courier/offers…` |
| **deliveries** | Active list, detail workflow, QR pickup, POD, failed | `/courier/deliveries…` |
| **navigation** | Map, polyline, ETA, external Maps | `GET …/route`, `POST /courier/location` |
| **shifts** | Start / end / break, overtime | `/courier/shifts` |
| **earnings** | Daily / weekly / monthly + payouts | `/courier/earnings` |
| **performance** | Acceptance, completion, on-time, rating, safety | `/courier/performance` |
| **profile** | Personal, vehicle, bank, tax; `DocumentsScreen` KYC | `/courier/profile` |
| **support** | AI chat, tickets, SOS (`tel:` + incident), reports | `/courier/support/*`, `/courier/incidents` |
| **notifications** | Inbox list | `/courier/notifications` |
| **settings** | Language, theme, biometric | local prefs |

## Providers

- `apiClientProvider`, `preferencesStoreProvider` — `lib/di/providers.dart`
- `courierSessionProvider` — alias of `authSessionProvider` in auth
- `dutyControllerProvider` — duty status with `DutyRules` validation

## Business rules

- `DutyRules` — duty status machine
- `DeliveryRules` — delivery lifecycle, pickup QR, POD, failure reasons
- `LocationRules` — ping intervals, low-battery throttle, impossible-jump detection

Unit tests: `test/unit/business_rules/`.

## Run

```bash
cd apps/mobile_courier
flutter pub get
flutter gen-l10n   # optional; screens use English literals until l10n is wired
flutter test test/unit/business_rules
flutter run
```

Override `preferencesStoreProvider` (and related stores) in bootstrap before `runApp`.

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for navigation graph, offline strategy, and duty/delivery state machines.
