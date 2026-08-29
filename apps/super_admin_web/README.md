# NEXORA Super Admin Web

Platform / multi-tenant control plane for the NEXORA quick-commerce ecosystem.

**Not** Admin Web city-ops. This app governs companies, tenants, licenses, flags, security, compliance, infrastructure, FinOps, AI platform, DR, deployments, and platform audit — never city order boards, courier dispatch, or warehouse pick UIs.

| | Super Admin (`super_admin_web`) | Admin (`admin_web`) |
|---|--------------------------------|---------------------|
| Scope | Platform / multi-tenant | City / company ops |
| Orders / live map | KPI aggregates only | Yes |
| Tenants | Create / isolate / suspend | Consume |
| Kill switches | Global + dual-control | Soft city flags |
| K8s / Kafka / DR | Yes | No |
| Port (dev) | typically `3001` if Admin uses `3000` | `3000` |

See [`ARCHITECTURE.md`](./ARCHITECTURE.md) and [`FEATURES.md`](./FEATURES.md).

## Prerequisites

- Node.js 20+
- Shared UI package at `packages/web/ui` (`@nexora/ui`)
- `identity-service` on `http://localhost:8081` for OTP login
- `platform-ops-service` on `http://localhost:8110` (fixture fallbacks only run when `NEXT_PUBLIC_ALLOW_MOCK_FALLBACK=true`)

## Setup

```bash
cd apps/super_admin_web
cp .env.example .env.local
npm install
```

`.env.local`:

```env
NEXT_PUBLIC_PLATFORM_OPS_URL=http://localhost:8110/v1
```

All platform API calls use `platformPath('/…')` → `/platform/…` under that base.

## Run

```bash
npm run dev
```

Open the URL Next prints (default `http://localhost:3000`). If Admin Web already owns `3000`:

```bash
npx next dev --turbopack -p 3001
```

Login is a real phone-number OTP against `identity-service`. Only an identity principal carrying the `super_admin` role is accepted; it is mapped to `platform_owner`. MFA and WebAuthn start unverified and are only flipped by the in-app step-up flows.

## Scripts

| Script | Purpose |
|--------|---------|
| `npm run dev` | Next.js dev (Turbopack) |
| `npm run build` / `npm start` | Production build & serve |
| `npm run lint` | ESLint |
| `npm test` | `vitest run` (unit tests) |
| `npm run test:watch` | Vitest watch |
| `npx tsc --noEmit` | Typecheck |

## Dual-control

Sensitive actions require a second distinct approver (`dual_control:approve`):

- Kill switches
- Tenant suspend / delete
- **DR failover**
- Secret rotation
- License overrides

See `src/shared/permissions/dual-control.ts`.

## Hard rules

- Dense NEXORA teal (`#0B6E6E`) via `@nexora/ui` tokens — no purple/cream themes
- Do **not** invent city order management, courier live maps, or CRM ticket inboxes
- Notifications here = **provider hub** (email/SMS/push/WhatsApp/webhooks), not city-ops inbox
