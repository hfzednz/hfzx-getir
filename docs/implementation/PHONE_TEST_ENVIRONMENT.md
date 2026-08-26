# Phone test environment — staging backend + Android APK (Prompt 64)

Native Flutter customer app (`apps/mobile_customer`). **No customer web app.** No Flutter Web.

## Result

**STAGING BACKEND — EXTERNAL CREDENTIAL REQUIRED**

A public HTTPS staging API was **not** deployed in this pass. No cloud/K8s/PaaS credentials are available in the environment or GitHub Actions (only Android keystore / Play secrets are referenced in CI).

Prepared for deployment:

| Artifact | Path |
|---|---|
| Disposable phone-test stack script | `scripts/staging/deploy-phone-test.sh` |
| Staging Android APK/AAB build | `scripts/ci/android-staging-customer.sh` |
| Manual CI workflow | `.github/workflows/cd-staging-android.yml` |

---

## STAGING API

```
(not deployed — provide cloud/VPS credentials and a real domain)
```

When deployed, the public URL must be supplied by the operator, e.g.:

`https://api-staging.<YOUR-DOMAIN>/v1`

Do **not** use `localhost`, `127.0.0.1`, `api.nexora.local`, or unresolvable `*.staging.nexora.io` placeholders.

---

## Deploy staging backend (operator steps)

### Option A — Docker VPS (smallest path, no K8s)

Requires: **Linux VPS with Docker**, **public IP**, **DNS A record**, optional **domain**.

```bash
# On the VPS (Ubuntu 22.04+ recommended)
git clone https://github.com/hfzednz/hfzx-getir.git
cd hfzx-getir
export STAGING_DOMAIN=api-staging.yourdomain.com
export STAGING_PUBLIC_URL=https://api-staging.yourdomain.com/v1
bash scripts/staging/deploy-phone-test.sh
```

What this runs:

| Component | Mode |
|---|---|
| PostgreSQL | **Not used** (in-memory dev mode — disposable) |
| Redis | **Not used** (in-memory dev mode) |
| Kafka | **Not used** (noop publisher in dev mode) |
| Services | identity, catalog, cart, location, checkout, payment, order, inventory, finance-ledger, settlement, **bff-customer** |
| Payment | **MockPSP** sandbox (`tok_ok` in API tests) |
| OTP | **`OTP_DEV_MODE=true`** — codes logged in identity container |

Verify OTP after login attempt:

```bash
docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode
```

HTTPS: script starts **Caddy** when `STAGING_DOMAIN` is set. Point DNS to the VPS before testing from a phone.

### Option B — Existing repo infra (K8s / Terraform / Helm)

Requires credentials **not present in repo**:

| Credential | Purpose |
|---|---|
| Cloud provider / `KUBECONFIG` | EKS/GKE/AKS or self-managed cluster |
| Terraform state backend | Apply `infra/terraform/envs/staging` |
| Container registry push | GHCR or cloud registry (CI builds images but does not push) |
| DNS + TLS | Replace `api.nexora.local` in Helm ingress |
| `DATABASE_URL` | Staging PostgreSQL (Neon, RDS, etc.) |
| `REDIS_URL` | Staging Redis (Upstash, ElastiCache, etc.) |
| `KAFKA_BROKERS` | Staging Kafka (Confluent, MSK, etc.) |
| `JWT_KEY_PEM`, `OTP_PEPPER` | Required when `DATABASE_URL` is set |
| SMS provider OR `OTP_DEV_MODE=true` on isolated staging only | Phone OTP |

Use `cd-gitops.yml` (manual dispatch, `environment: staging`) after images and cluster exist.

---

## ANDROID APK

### Local build (this audit)

| Field | Value |
|---|---|
| Path | `apps/mobile_customer/build/app/outputs/flutter-apk/app-release.apk` |
| Size | ~89.7 MB |
| Signing | **Debug-signed release** (no `ANDROID_KEYSTORE_*` secrets) |
| API configured | Placeholder `https://STAGING-API-URL-REQUIRED/v1` — **rebuild after backend is live** |

Rebuild with real staging URL:

```bash
export NEXORA_STAGING_BASE_URL=https://api-staging.yourdomain.com/v1
bash scripts/ci/android-staging-customer.sh
```

### CI artifact (recommended after backend is live)

1. GitHub → Actions → **cd-staging-android** → Run workflow
2. Input `staging_api_url`: `https://api-staging.yourdomain.com/v1`
3. Download artifacts: `customer-staging-apk`, `customer-staging-aab`

Signing: release keystore if GitHub secrets set; otherwise debug-signed (installable on device with “unknown sources”).

---

## ANDROID AAB

Same build script produces AAB at:

`apps/mobile_customer/build/app/outputs/bundle/release/app-release.aab`

Use Play **Internal testing** when `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON` and keystore secrets are configured (`cd-mobile.yml`).

---

## TEST ACCOUNT

| Field | Value |
|---|---|
| Tenant | App sends tenant context; CI uses `11111111-1111-1111-1111-111111111111` |
| Phone | `+905551112233` |
| OTP | Read from identity logs when `OTP_DEV_MODE=true` (deploy-phone-test stack) |
| Negative phone | `+905551119999` (CI negative tests) |

Do **not** use production customer credentials.

---

## PAYMENT MODE

**SANDBOX** — MockPSP in disposable staging stack. No real charges.

Payment token in API tests: `tok_ok`. Production `STRIPE_SECRET_KEY` must **not** be used on disposable staging unless explicitly intended as Stripe **test** mode.

---

## INSTALL (Android phone)

1. Deploy staging backend and confirm `curl https://api-staging.yourdomain.com/health` returns 200.
2. Build or download staging APK with **matching** `NEXORA_STAGING_BASE_URL`.
3. Transfer APK to phone (USB, Drive, CI artifact download).
4. Enable **Install unknown apps** for the file source.
5. Install `app-release.apk`.
6. Open app → enter test phone → read OTP from server logs → complete checkout.

---

## CUSTOMER CHECKOUT STEPS (phone)

1. **Install** staging APK  
2. **Open** app (splash / NEXORA)  
3. **Login** — phone `+905551112233`, OTP from logs  
4. **Location** — grant permission or enter address  
5. **Home** — catalog tiles load from staging BFF  
6. **Category / Product** — browse and open detail  
7. **Add to cart**  
8. **Cart** — adjust quantity  
9. **Checkout** — confirm address / delivery  
10. **Payment** — MockPSP sandbox  
11. **Order confirmation**  
12. **Order history / detail**

Backend verification after order:

```bash
# BFF health + request IDs
curl -sS -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  https://api-staging.yourdomain.com/v1/customer/home?lat=41.0&lng=29.0

# Service logs
docker logs nexora-staging-bff-customer --tail 50
docker logs nexora-staging-order-service --tail 50
docker logs nexora-staging-payment-service --tail 50
```

---

## Status matrix (this pass)

| Gate | Status |
|---|---|
| STAGING API (public HTTPS) | **BLOCKED** — no cloud/VPS credentials |
| BACKEND services | **BLOCKED** — not deployed |
| DATABASE (PostgreSQL) | **BLOCKED** — disposable stack uses in-memory mode |
| REDIS | **BLOCKED** — not provisioned (in-memory mode) |
| KAFKA | **BLOCKED** — not provisioned (noop in dev mode) |
| ANDROID APK (staging config) | **READY** — build script + local APK produced |
| ANDROID AAB (staging config) | **READY** — same script |
| CHECKOUT on phone | **BLOCKED** — requires live staging API + rebuilt APK |
| STORE SIGNING | **BLOCKED** — no keystore secrets |

---

## Known limitations

1. **No public staging API** until operator provisions VPS/cloud + domain.
2. **Disposable in-memory stack** — data resets on container restart; not suitable for soak tests. For persistent staging, provision PostgreSQL/Redis/Kafka and set `DATABASE_URL` (+ auth secrets).
3. **APK built with placeholder URL** in this audit — must rebuild after staging API is live.
4. **Maps** use placeholder key unless `GOOGLE_MAPS_API_KEY` is set at build time.
5. **Physical phone test not executed** here — no live backend to connect to.
6. **iOS** not covered — use RC unsigned iOS build + TestFlight when Apple credentials exist.

---

## Repository

| Field | Value |
|---|---|
| Commit (doc) | see git log on `main` |
| Customer app | `apps/mobile_customer` |
| Application ID | `com.hfzx.nexora.nexora_customer` |
| CI unchanged | `ci-acceptance`, `ci-release-candidate`, `ci-quality` — no modifications to existing jobs |

## Related

- `docs/implementation/MOBILE_TEST_ENVIRONMENT.md` — Prompt 63 (no web URL)
- `docs/implementation/FINAL_PRODUCTION_READINESS.md` — STATE B certification
- `scripts/ci/e2e-smoke.sh` — checkout dependency chain reference
