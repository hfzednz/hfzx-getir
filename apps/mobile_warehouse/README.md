# NEXORA Warehouse (Dark Store)

Dense Flutter ops client for pick → pack → handoff, inventory, and dark-store workflows.

Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md) · Design density: **Dense** (`NxDensity.dense`)

## Stack

| Concern | Choice |
|---------|--------|
| State | Riverpod |
| Routing | GoRouter (`lib/routing/app_router.dart`) |
| HTTP | Dio via `ApiClient` (`apiClientProvider`) |
| Session | `warehouseSessionProvider` (+ `storeIdProvider`) |
| Realtime | `RealtimeClient` — `handoff.*` invalidates dispatch queue |
| Scan | `mobile_scanner` |
| Local | Hive `PreferencesStore` + mutation outbox |

## DI providers

- `apiClientProvider`, `preferencesStoreProvider` — `lib/di/providers.dart`
- `storeIdProvider` — derived from session
- `warehouseSessionProvider` — alias over session notifier (swap to auth session when OTP lands)

Override `preferencesStoreProvider` / `mutationOutboxProvider` in bootstrap before `runApp`.

## Modules

See [`FEATURES.md`](FEATURES.md).

## Tests

```bash
cd apps/mobile_warehouse
flutter test test/unit/business_rules
```

Covers `PickingRules`, `PackingRules`, `InventoryRules`, `HandoffRules`, `RoleRules`.

## Run

```bash
cd apps/mobile_warehouse
flutter pub get
flutter run --dart-define=NEXORA_BASE_URL=… --dart-define=NEXORA_WS_URL=…
```

Platform folders (`android`/`ios`) may still need `flutter create .` if not scaffolded yet.
