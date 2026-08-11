# NEXORA Admin Web — Feature Status

> Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md) · Density: **Dense** · UI: `@nexora/ui`

| Module | Route(s) | Status |
|--------|----------|--------|
| Auth & Session | `/login` | Implemented (demo session; OIDC via bff-admin next) |
| Shell | `(app)` layout | Implemented (nav, city switcher, ⌘K, theme) |
| Dashboard | `/dashboard` | Implemented |
| Live Operations | `/live` | Implemented (+ WS→poll fallback) |
| Orders | `/orders`, `/orders/[id]` | Implemented (RBAC-gated interventions) |
| Customers | `/customers`, `/customers/[id]` | Implemented (360 profile) |
| Couriers | `/couriers`, `/couriers/[id]` | Implemented |
| Warehouses | `/warehouses`, `/warehouses/[id]` | Implemented |
| Products | `/products`, `/products/[id]`, `/products/import` | Implemented |
| Inventory | `/inventory` | Implemented |
| Delivery | `/delivery`, `/delivery/zones` | Implemented |
| Campaigns | `/campaigns`, `/campaigns/[id]` | Implemented |
| Pricing | `/pricing` | Implemented |
| Loyalty | `/loyalty` | Implemented |
| CRM | `/crm` | Implemented |
| Support | `/support`, `/support/tickets/[id]` | Implemented |
| Finance | `/finance` | Implemented (payout gates) |
| Analytics | `/analytics` | Implemented |
| AI Command | `/ai` | Implemented |
| System | `/system`, `/flags`, `/templates` | Implemented (kill switches → super_admin) |
| RBAC | `/rbac` | Implemented |
| Audit | `/audit` | Implemented |
| Notifications | `/notifications` | Implemented |
| Monitoring | `/monitoring` | Implemented |
| Reports | `/reports` | Implemented (mock export) |

## Tests

`npm test` — permission + order-rules unit tests (vitest).

## BFF

`NEXT_PUBLIC_BFF_ADMIN_URL` (default `http://localhost:8084/v1`). Features call BFF then fall back to mock data so the UI remains operable offline.
