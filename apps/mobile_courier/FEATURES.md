# Courier App — Feature Status (Prompt 05)

> Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md) · Design density: **Compact**

## Modules

| Module | Status |
|--------|--------|
| Auth OTP + KYC gate | Implemented |
| Duty status (online/offline/busy/break/emergency) | Implemented + `DutyRules` |
| Home dashboard | Implemented |
| Offers accept/reject + realtime | Implemented |
| Deliveries workflow + QR pickup + POD + failed | Implemented + `DeliveryRules` |
| Live navigation map + external nav | Implemented |
| Location adaptive pings + spoof hooks | Implemented + `LocationRules` |
| Shifts | Implemented |
| Earnings | Implemented |
| Performance | Implemented |
| Profile + documents | Implemented |
| Support + SOS/incidents | Implemented |
| Notifications + settings | Implemented |
| Offline Hive local store + outbox | Implemented |
| Analytics event constants | Implemented |
| EN/TR l10n ARB | Present |

## Tests

`test/unit/business_rules/` — duty, delivery, location (**22 passing**).

## Run

```bash
cd apps/mobile_courier
flutter pub get
flutter run --dart-define=NEXORA_BASE_URL=… --dart-define=NEXORA_WS_URL=…
```
