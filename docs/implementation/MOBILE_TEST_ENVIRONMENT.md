# Mobile test environment — staging URL status (Prompt 63)

Authoritative record for human mobile-browser testing of the NEXORA customer product. No fabricated URLs or credentials.

## Result

**STAGING DEPLOYMENT — EXTERNAL CREDENTIAL REQUIRED**

A public customer test URL was **not** created in this pass because:

1. The customer product is **Flutter mobile-only** (`apps/mobile_customer`) — there is no customer web app in the repository.
2. An experimental `flutter build web` **failed** (Drift/SQLite FFI is not web-compatible). Enabling mobile-browser testing requires either a dedicated web client or native app distribution — not a one-line deploy of the existing mobile binary.
3. Documented staging hostnames (`api.staging.nexora.io`, `app.staging.nexora.io`) **do not resolve** on the public internet (DNS probe failed 2026-08-26).
4. No live staging cluster, Vercel project, or GitHub Pages customer deployment exists for this repository.
5. Deployment credentials are **not** available in GitHub Actions (only Android keystore / Play secrets are referenced in `cd-mobile.yml`).
6. Local agent has **no Docker**, **no tunnel tools**, and **no cloud CLI** to stand up a persistent public stack.

## Customer application (verified)

| Field | Value |
|---|---|
| Path | `apps/mobile_customer` |
| Package | `nexora_customer` v1.0.0+1 |
| Type | **Flutter mobile** (Android / iOS) |
| Web | **Not supported** (no pre-existing `web/` target; build blocked by `drift` / `sqlite3` FFI) |
| Other apps | `admin_web`, `super_admin_web` are **operator** dashboards — not customer UX |

## Documented but non-live endpoints

| Host | Purpose | Public DNS (2026-08-26) |
|---|---|---|
| `https://api.staging.nexora.io/v1` | Staging BFF/API (code preset in `nexora_core`) | **NX / unreachable** |
| `wss://realtime.staging.nexora.io/v1` | Staging WebSocket | **NX / unreachable** |
| `https://api.nexora.io/v1` | Production API placeholder | **Not verified live** |
| `api.nexora.local` | Helm ingress default (`values.yaml`) | Internal K8s only |
| `api.nexora.example` | Prod Helm placeholder (`values/prod.yaml`) | Documentation only |

## Existing deployment machinery (not provisioned)

| Mechanism | Location | Status |
|---|---|---|
| Terraform staging | `infra/terraform/envs/staging` | Module definitions only — no applied state in repo |
| K8s staging overlay | `infra/k8s/overlays/staging` | Image tag patch only |
| Helm chart | `infra/helm/nexora` | `ingress.hosts` = `api.nexora.local` |
| Argo CD ApplicationSet | `infra/argocd/applicationset.yaml` | Targets `https://github.com/nexora/platform.git` (placeholder org repo) |
| GitOps promote | `.github/workflows/cd-gitops.yml` | Manual dispatch — opens PR for image tag bump |
| Mobile store CD | `.github/workflows/cd-mobile.yml` | Requires keystore / Play secrets |
| Vercel | — | **Not configured** |
| GitHub Pages | — | **Not deployed** (`hfzednz.github.io/hfzx-getir/` → 404) |

## What a working staging stack needs

### Frontend (mobile browser requirement)

Choose **one**:

| Option | Requirement |
|---|---|
| **A. Native app (recommended for current codebase)** | Install debug/signed APK or TestFlight build from RC artifact / `cd-mobile.yml`; point app at staging BFF via `--dart-define=NEXORA_BASE_URL=...` |
| **B. Flutter web** | Refactor offline storage (Drift → web-safe store), stub/replace mobile-only plugins (maps, scanner, biometrics), add `web/` target, deploy static host |
| **C. Separate customer web app** | New Next.js/React storefront — out of scope for “do not redesign” |

### Backend (required for any client)

Deploy to a public host (examples aligned with repo infra):

- `bff-customer` (+ `identity-service`, `catalog-service`, `cart-service`, `location-service`, minimum)
- For full checkout: add `checkout-service`, `payment-service`, `order-service`, `inventory-service`, `finance-ledger-service`, `settlement-service`
- PostgreSQL, Redis, Kafka per `infra/docker/docker-compose.yml` or managed equivalents
- Ingress TLS hostname e.g. `api.staging.<your-domain>`

### Secrets / credentials required (exact gaps)

| # | Credential / input | Used for |
|---|---|---|
| 1 | **Cloud/K8s access** — e.g. `KUBECONFIG`, EKS/GKE/AKS credentials, or PaaS token (Render, Railway, Fly.io) | Run staging backend + ingress |
| 2 | **Public DNS + TLS** — domain + cert manager or platform TLS | `https://api.staging.<domain>` |
| 3 | **Database URL** — staging PostgreSQL (not production) | Migrations + seed data |
| 4 | **Redis / Kafka URLs** — staging instances | Cart, events, recovery paths |
| 5 | **JWT_KEY_PEM**, **OTP_PEPPER** (staging values) | Auth |
| 6 | **STRIPE_SECRET_KEY** or keep **MockPSP** | Payment (`MockPSP` OK for staging; forbidden in prod) |
| 7 | **CORS allowlist** including customer web origin if option B/C | Browser API calls |
| 8 | Optional: **GOOGLE_MAPS_API_KEY**, Firebase for maps/push on device | Full mobile UX |

No item above is committed to git (correct).

## Test account method (when backend is live)

From CI harness (`scripts/ci/e2e-smoke.sh`, Playwright `customer.journey.spec.ts`):

| Field | Value |
|---|---|
| Tenant header | `X-Tenant-Id: 11111111-1111-1111-1111-111111111111` |
| Test phone | `+905551112233` |
| OTP | With `OTP_DEV_MODE=true` and empty `DATABASE_URL`, identity logs OTP codes (dev/in-memory mode) |
| Staging rule | Use **`OTP_DEV_MODE=true` only on staging**; must be **`false` in production** (`infra/helm/nexora/values/prod.yaml`) |

## Payment mode (staging)

| Mode | When |
|---|---|
| **MockPSP / sandbox** | Default in local/CI disposable harness — no real charges |
| **Stripe test keys** | When `STRIPE_SECRET_KEY` set to test mode |
| **Production PSP** | **Forbidden** for staging/test per `docs/production/ENVIRONMENTS.md` |

## Smallest path to a real TEST URL (human checklist)

1. Provision staging cluster or PaaS project using `infra/terraform/envs/staging` + Helm/Compose as reference.
2. Set ingress to a **real domain** (replace `api.nexora.local` / `api.nexora.example`).
3. Run migrations; load synthetic catalog/inventory seed (see service migrations under `services/*/migrations/`).
4. Deploy `bff-customer` and dependencies with `OTP_DEV_MODE=true`, MockPSP, staging CORS.
5. **For mobile browser:** either complete Flutter web work (option B) **or** distribute native build and document deep link / custom URL scheme (`nexora://`) — browser URL requires web client.
6. Record the live URL in this document and re-run Prompt 63 validation matrix.

## Validation matrix (this pass)

| Check | Result | Notes |
|---|---|---|
| HTTPS public URL | **FAIL** | No deployment |
| DNS | **FAIL** | `*.staging.nexora.io` does not resolve |
| Frontend loads | **FAIL** | No customer web host |
| API reachable | **FAIL** | No public staging API |
| Auth | **BLOCKED** | Depends on backend |
| Catalog / cart / checkout | **BLOCKED** | Depends on backend + client |
| Mobile responsive web | **FAIL** | No web client |
| Physical phone test | **NOT EXECUTED** | No URL to open |

## Repository state

| Field | Value |
|---|---|
| Branch | `main` |
| Commit audited | `cad8885d59ca8ea107e8ca85b544aeb6e2f3d4c2` (Prompt 62 final doc) |
| CI status | Green on `23c1225` / `cad8885` lineage — **unchanged by this audit** |
| Code changes in Prompt 63 | **None committed** (experimental web scaffold reverted) |

## Related documents

- `docs/implementation/FINAL_PRODUCTION_READINESS.md` — STATE B certification
- `store/EXTERNAL_INPUTS.md` — store/legal external inputs
- `docs/production/ENVIRONMENTS.md` — environment matrix
- `docs/launch/DEPLOYMENT_GUIDE.md` — GitOps deploy flow
- `apps/mobile_customer/README.md` — native run instructions
