# NEXORA Courier App — Architecture

> Binding under Master Blueprint + Design System.  
> Density: **Compact** (DS §00). Shares `nexora_core` + `nexora_design`.  
> BFF: `bff-courier` at `/v1`.

## Stack

| Concern | Choice |
|---------|--------|
| Flutter | Stable |
| State | Riverpod |
| Routing | GoRouter |
| HTTP | Dio via `ApiClient` |
| Local | Drift + Hive outbox |
| Maps | google_maps_flutter (+ url_launcher to Apple/Google Maps nav) |
| Location | Geolocator + background service hooks |
| Push | FCM |
| Realtime | WebSocket `RealtimeClient` |
| Auth | OTP + biometric + secure storage |
| Scan | mobile_scanner (warehouse QR / package) |

## Folder structure

```text
apps/mobile_courier/lib/
  main.dart
  bootstrap/
  app/
  di/
  routing/
  l10n/
  shared/            # validators, business_rules, analytics, widgets, location
  features/
    splash/
    auth/            # OTP, KYC, documents, vehicle
    home/            # dashboard
    status/          # online/offline/busy/break/emergency
    offers/          # incoming assignment offers
    deliveries/      # active jobs, batch, POD
    navigation/      # route, ETA, map chrome
    shifts/
    earnings/
    performance/
    profile/
    support/
    notifications/
    settings/
    shell/
```

Each feature: `domain/` · `data/` · `presentation/` (providers, screens, widgets).

## Navigation graph

```text
/                         splash
/auth                     phone → otp → kyc gate
/shell
  /home                   dashboard
  /offers                 incoming + history
  /deliveries             active list
  /earnings
  /account
/deliveries/:id           detail + workflow
/deliveries/:id/navigate  live map
/deliveries/:id/pickup    warehouse scan
/deliveries/:id/pod       proof of delivery
/shifts
/performance
/profile/**
/support/**
/settings/**
/notifications
```

Protected: all shell + delivery paths require authenticated courier with KYC approved (soft-gate: KYC incomplete → docs screens).

## State management

- Global: `apiClient`, `courierSession`, `dutyStatus`, `cityId`, `theme`, `locale`
- Feature Notifiers for duty status, active delivery, offer queue
- Drift: offline delivery tasks + location breadcrumbs + outbox mutations
- Realtime: offer push + assignment updates invalidate providers

## Duty status machine

```text
offline → online → (busy | on_break | emergency)
online → assigned → en_route_store → at_store → picked_up → en_route_customer → arrived → delivered|failed
```

Illegal transitions rejected by `DutyRules` / `DeliveryRules`.

## Offline-first

| Domain | Strategy |
|--------|----------|
| Active delivery snapshot | Drift SoR |
| Status transitions | Outbox + server state machine |
| Location pings | Batch queue; throttle on low battery |
| Map tiles | OS/SDK cache (no custom tile server required) |
| Accept/reject | Online preferred; queue only if reconnect within TTL |

## Battery / location

- Adaptive interval: online idle 30–60s; active delivery 3–8s; low battery 60–120s
- Foreground service while `online` or active delivery (platform channel / plugin hooks)
- Spoofing hooks via `DeviceIntegrityChecker` + impossible jump detection in `LocationRules`

## Dependency graph

```text
UI → Riverpod → Use cases → Repositories → Remote (bff-courier) | Local (Drift/Hive)
nexora_design ← presentation
nexora_core ← network, sync, security
```

## Admin / warehouse integration

- Status + location streamed to dispatch via Kafka consumers (server)
- Pickup QR verifies `warehouse_handoff` token from `warehouse-service`
- Incidents POST `/courier/incidents` → CRM/ops

## Modules ↔ Blueprint services

| App module | Primary APIs |
|------------|--------------|
| Auth/KYC | identity + courier profile |
| Offers/Deliveries | dispatch + order |
| Navigation | routing + tracking ingest |
| Earnings | settlement |
| Shifts | workforce/config |
