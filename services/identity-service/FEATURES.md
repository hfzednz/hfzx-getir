# Identity Service — Feature Status

| Feature | Status | Notes |
|---------|--------|-------|
| Domain model (principal, session, device, RBAC, tokens, audit, risk) | **Done** | `internal/domain` |
| Security: JWT RS256 + JWKS | **Done** | `internal/security/jwt` |
| Security: password / OTP / TOTP / refresh / WebAuthn helpers | **Done** | packages under `internal/security` |
| Risk scoring helpers | **Done** | `internal/risk` |
| Rate limit (memory + Redis interface) | **Done** | `internal/ratelimit` |
| Env config | **Done** | `internal/config` |
| App ports (hexagonal repositories) | **Done** | `internal/app/ports` |
| App use cases (OTP, password, magic, social, WebAuthn, guest, MFA, sessions, devices, RBAC PDP, privacy) | **Done** | `internal/app` on `Deps` |
| In-memory repos (dev / tests) | **Done** | `internal/app/memory` |
| HTTP REST (ARCHITECTURE routes) | **Done** | stdlib ServeMux Go 1.22+ → `Deps` |
| HTTP middleware (requestID, log, recover, CORS, rate limit) | **Done** | |
| NEXORA JSON error envelope | **Done** | camelCase `error.{code,message,traceId,retriable}` |
| OIDC discovery / authorize / token / JWKS / userinfo | **Partial** | Code+PKCE in-memory store; production-shaped |
| Postgres adapter | **Partial** | Compiling stubs implementing ports; needs pgx driver for live DB |
| Redis session cache | **Stub** | In-process map until go-redis wired |
| Kafka publisher | **Stub** | Logs / buffers; implements `ports.EventPublisher` |
| gRPC CheckPermission / ValidateSession | **Stub** | Hand-written types; proto present, no codegen |
| OpenAPI 3.1 | **Done** | `api/openapi/identity-v1.yaml` |
| Proto identity.v1 | **Done** | `proto/identity/v1/identity.proto` |
| Docker / compose / Makefile / README | **Done** | |
| Password verify + session mint + refresh rotation | **Done** | App layer + `DefaultTokenIssuer` |
| WebAuthn ceremony (stub service) | **Partial** | Interfaces wired; real FIDO2 library TBD |
| Social IdP federation | **Partial** | Stub IdPs; callback path ready |
| ABAC/RBAC PDP evaluation | **Done** | `CheckPermission` / `ListEffectivePermissions` |
| Audit / risk Kafka topics production client | **TODO** | Publisher stub only |

Legend: **Done** = usable · **Partial** = works for compile/demo · **Stub** = interface only · **TODO** = not started
