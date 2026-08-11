# NEXORA Identity Service

Go microservice for authentication, authorization (RBAC/ABAC PDP), sessions, devices, MFA, and OIDC.

Module: `github.com/nexora/identity-service`

## Quick start (dev / in-memory)

```bash
cd services/identity-service
go run ./cmd/identity-service
```

With empty `DATABASE_URL` the process boots in **dev mode** using in-memory repositories. OTP codes are logged when `OTP_DEV_MODE=true` (default).

Health:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8080` | Listen address |
| `DATABASE_URL` | empty | Postgres DSN; empty → memory repos |
| `REDIS_URL` | empty | Optional Redis; when set, SessionCache is pinged for readiness and closed on shutdown. Session **persistence** uses `SessionRepository` (Postgres or in-memory DevMode) — Redis is not the session SoT |
| `KAFKA_BROKERS` | empty | Comma-separated brokers (stub publisher) |
| `JWT_ISS` / `JWT_AUD` | local defaults | Access token issuer / audience |
| `JWT_KEY_PEM` | empty | RSA PEM path; empty → ephemeral key |
| `OTP_DEV_MODE` | `true` | Log OTP / magic-link codes |
| `PUBLIC_BASE_URL` / `OIDC_ISSUER` | `http://localhost:8080` | Public OIDC issuer |
| `CORS_ALLOWED_ORIGINS` | `*` | CORS |
| `RATE_LIMIT_PER_MINUTE` | `120` | Per-IP rate limit |

See `configs/config.example.yaml`.

## Docker Compose

```bash
docker compose up --build -d postgres redis identity-service
# optional Kafka (Redpanda):
docker compose --profile kafka up -d redpanda
```

Migrations in `migrations/` are mounted into Postgres init.

## Auth flows (summary)

1. **Phone OTP** — `POST /v1/identity/auth/otp/start` → verify → tokens  
2. **Password** — `POST /v1/identity/auth/password/login`  
3. **Magic link** — start / consume  
4. **Social** — start / callback stubs  
5. **WebAuthn** — register & authenticate option/verify stubs  
6. **Guest** — limited session  
7. **MFA** — TOTP enroll, challenge, verify  
8. **Tokens** — refresh / revoke / introspect  
9. **OIDC** — discovery, authorize (code+PKCE), token, JWKS, userinfo  

OpenAPI: `api/openapi/identity-v1.yaml`  
gRPC proto: `proto/identity/v1/identity.proto` (`CheckPermission`, `ValidateSession`)

## Makefile

```bash
make test
make run
make build
make migrate   # requires DATABASE_URL + psql
```

## Layout

```
cmd/identity-service/     main
internal/config/          env config
internal/app/             use cases + ports
internal/adapters/http/   REST + OIDC
internal/adapters/postgres|redis|kafka|grpc|memory
internal/domain|security|risk|ratelimit
migrations/ api/openapi/ proto/
```

## Feature status

See [FEATURES.md](./FEATURES.md).
