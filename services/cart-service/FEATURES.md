# cart-service features

## Implemented

| Area | Status | Notes |
|------|--------|-------|
| Create / get cart (guest or principal) | ✅ | Idempotent get-or-create per owner |
| Add / update / remove line | ✅ | Opaque `variant_id`; qty sum on add |
| Max qty (available) at line level | ✅ | `max_qty`; `ErrMaxQtyExceeded` |
| Apply / remove coupon | ✅ | Preview codes; pricing owns discount SoT |
| Refresh quote | ✅ | `PricingClient.Quote` → snapshot (minor units) |
| Soft reserve on refresh (optional) | ✅ | `softReserve: true` |
| SoftReserveLines | ✅ | `InventoryClient.ATP` + SoftReserve/Release |
| Merge carts on login | ✅ | Default policy: qty sum |
| Mark abandoned / recover | ✅ | Status transitions |
| Recommendations | ✅ | `RecommendClient` stub |
| Save cart (save-for-later) | ✅ | Requires `X-Nexora-User` |
| Outbox + cart.lifecycle events | ✅ | Memory + Kafka stub |
| HTTP `:8087` `/v1/cart/...` | ✅ | NEXORA errors, tenant/guest/user headers |
| Migrations 001–010 | ✅ | carts→indexes |
| OpenAPI + proto | ✅ | Contract stubs |
| Docker / Makefile | ✅ | Dev compose |

## Partial / stubs

| Area | Status | Notes |
|------|--------|-------|
| Postgres adapters | 🔶 | Schema ready; runtime memory until wired |
| Kafka publish | 🔶 | Stub logs when brokers set |
| Pricing / inventory HTTP | 🔶 | In-process memory stubs |
| gRPC | 🔶 | Proto defined; listener stub |
| Redis | 🔶 | Client stub |

## Non-goals

- ❌ Product content / catalog ownership
- ❌ Stock ledger / ATP SoT (calls inventory only)
- ❌ Order aggregate / place saga
- ❌ PSP charges / payment capture
- ❌ Float money
