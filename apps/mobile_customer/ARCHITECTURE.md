# NEXORA Customer Mobile App — Architecture

> Binding under Master Blueprint + Design System (`docs/design-system/00-INDEX.md`).  
> App package: `nexora_customer` · Shared: `nexora_core`, `nexora_design`

## Stack

| Concern | Choice |
|---------|--------|
| Framework | Flutter stable |
| State | Riverpod 2.x |
| Routing | GoRouter (nested shell + deep links) |
| HTTP | Dio via `nexora_core` `ApiClient` |
| Local DB | Drift (SQLite) |
| KV / outbox | Hive CE |
| Secure tokens | Flutter Secure Storage |
| Push | Firebase Messaging |
| Maps | google_maps_flutter (Apple Maps via platform on iOS when configured) |
| Realtime | WebSocket `RealtimeClient` |
| Auth extras | Google / Apple / local_auth |
| Design | `nexora_design` only — no raw Material chrome |

## Folder structure

```text
apps/mobile_customer/lib/
  main.dart
  bootstrap/          # env, Hive, Drift, FCM, ProviderScope overrides
  app/                # NexoraApp (theme, l10n, router)
  di/                 # root providers
  routing/            # GoRouter, guards, route names
  l10n/               # ARB + generated
  shared/             # validators, money, business rules, widgets
  features/<feature>/
    domain/           # entities, repository ports, use cases
    data/             # datasources, models, repository impls, local
    presentation/     # providers, screens, widgets
```

## Feature modules

Auth · Onboarding · Splash · Home · Search · Categories · Product · Cart · Checkout ·
Orders · Tracking · Favorites · Wallet · Coupons · Loyalty · Notifications · Profile ·
Addresses · Reviews · Support · Settings · Referral · Help · About · Legal · City · AI · Shell

## State management

- Feature `*RepositoryProvider` → use cases → `FutureProvider` / `Notifier`
- Global: `apiClientProvider`, `themeModeProvider`, `localeCodeProvider`, `cityIdProvider`, session
- Checkout: `CheckoutController` (Notifier) holds address, schedule, payment, quote
- Tracking: `StreamProvider` over REST snapshot + realtime events
- Rebuild isolation: `select` / family providers; no god-providers

## Navigation graph

```text
/ (splash)
/onboarding
/auth/** (welcome, phone, otp, email, forgot, reset)
Shell tabs:
  /home
  /categories → /categories/:categoryId
  /search (+ /search/barcode)
  /cart
  /account
Root overlays:
  /p/:productId
  /checkout/{address|schedule|payment|review}
  /orders → /orders/:id → track | review
  /favorites /wallet /coupons /loyalty /notifications /profile
  /addresses[/add|/:id/edit]
  /support[/tickets/:id|/assistant]
  /settings/** (language, theme, a11y, notifications, privacy, devices, delete)
  /referral /help /about /legal/:doc
  /c/:slug /promo/:code
```

Protected routes: see `authRequiredPaths` in `routing/route_names.dart`.

## Dependency graph

```text
UI (presentation)
  → Riverpod providers
    → Use cases (domain)
      → Repositories (ports)
        → Remote DS (ApiClient) | Local DS (Drift/Hive)
nexora_design ← presentation only
nexora_core ← data + di + bootstrap
```

## Offline-first

| Domain | Strategy |
|--------|----------|
| Favorites | Drift SoR + pendingSync + outbox flush |
| Cart | Drift cache + optimistic qty; confirm online for checkout |
| Recent searches | Drift |
| Orders list | Cache for read; mutations online |
| Checkout / payment | Online-only |
| Tracking | Online + WS; last snapshot cached |

Conflict: server version wins; illegal transitions rejected; UI rebases.

## Security

- Certificate pinning adapter (core)
- Tokens in secure storage; refresh rotation
- Biometric gate optional
- Device integrity hooks (core)
- Idempotency-Key on checkout confirm

## Performance

- Lazy routes / autoDispose providers
- Pagination on category/search/orders
- Cached network images
- Skeleton loading (design system)
- Signature motions only (DS §14)

## Testing

- Unit: entities, business rules, validators
- Widget: smoke + critical screens
- Integration: boot → splash
- Golden: design package
