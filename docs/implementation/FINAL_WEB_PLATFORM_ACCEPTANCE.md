# Final web platform acceptance (Prompt 66)

**Status:** STATE B — technically complete; external deployment required

**Prior baseline:** `cdda568` (Prompt 65)

---

## Summary

Prompt 66 replaces demo staff logins with real identity OTP, wires staff dashboards to live BFF/service APIs where available, adds shared `@nexora/web-core` auth helpers, tenant-isolation Playwright tests, and CI web E2E against the disposable phone-test stack.

**Admin and super-admin apps remain demo-auth with mock-heavy APIs** — not upgraded in this prompt.

---

## Public URL matrix

| Application | URL | Status |
|-------------|-----|--------|
| Customer | — | BLOCKED (`VERCEL_TOKEN`) |
| Courier | — | BLOCKED |
| Warehouse | — | BLOCKED |
| Supplier | — | BLOCKED |
| Operations | — | BLOCKED |
| Support | — | BLOCKED |
| Finance | — | BLOCKED |
| Admin | — | BLOCKED |
| Super Admin | — | BLOCKED |

---

## Role matrix (honest)

| Role | App | Auth | Real API | Status |
|------|-----|------|----------|--------|
| customer | customer-web | OTP/BFF | Yes | Integrated |
| courier | courier-web | OTP/identity | Yes | Integrated |
| picker/packer/dispatcher | warehouse-web | OTP/identity | Yes | Integrated |
| supplier/partner | supplier-web | OTP/identity | Yes | Partial |
| city_ops | operations-web | OTP/identity | bff-admin (minimal) | Partial |
| support_agent | support-web | OTP/identity | bff-admin (minimal) | Partial |
| finance_analyst | finance-web | OTP/identity | Yes | Integrated |
| admin | admin_web | **Demo** | Partial mocks | Partial |
| super_admin | super_admin_web | **Demo** | Mock-heavy | Partial |

Detail: [WEB_ROLE_MATRIX.md](./WEB_ROLE_MATRIX.md)

---

## Deliverables

- `@nexora/web-core`: `otp-flow.ts`, `OtpLoginForm`, `RouteGuard`, `identityUrl()`
- Role apps: real OTP login + dashboard API loads
- Customer: real OTP; demo product fallback removed
- `scripts/ci/web-e2e.sh` + `tenant.isolation.spec.ts`
- `ci-web-quality.yml` web-e2e job
- `cd-web-staging.yml` (VERCEL gated, no auto-deploy)

---

## External inputs required

| Input | Purpose |
|-------|---------|
| `VERCEL_TOKEN` | Public web staging |
| Public BFF URLs | `NEXT_PUBLIC_BFF_*` |
| `NEXT_PUBLIC_IDENTITY_URL` | Staff OTP |
| DNS/TLS | HTTPS staging host |

---

## Known limitations

1. Multi-role E2E chain not automated in one Playwright suite
2. Realtime SSE not wired to all UIs (customer tracking polls)
3. `bff-admin` exposes ~6 live routes vs large admin UI surface
4. CI web E2E runs on GitHub Actions (Docker); not verified locally without Docker
