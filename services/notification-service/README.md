# Notification Service

Omnichannel messaging for NEXORA: templates, preferences/consent, scheduling, provider dispatch (FCM/APNs/SMTP/SMS/WhatsApp), inbox, retries/DLQ.

**Hard rules:** Does not own CRM tickets, campaign/coupon definitions, or order aggregates. Uses opaque `principal_id` / `order_id`.

## Quick start

```bash
cd services/notification-service
go test ./...
go build ./...
HTTP_ADDR=:8101 go run ./cmd/notification-service
```

Empty `DATABASE_URL` → in-memory repositories + mock providers (default).

## Environment

| Variable | Default | Notes |
|----------|---------|-------|
| `HTTP_ADDR` | `:8101` | REST |
| `GRPC_ADDR` | `:9101` | Stub |
| `DATABASE_URL` | _(empty)_ | DevMode when empty |
| `REDIS_URL` | _(empty)_ | Stub |
| `KAFKA_BROKERS` | _(empty)_ | Noop publisher |
| `RATE_LIMIT_PER_MINUTE` | `240` | Per IP |
| `CORS_ALLOWED_ORIGINS` | `*` | CSV |

## API

Base path: `/v1/notifications/...`

Headers: `X-Tenant-Id` (required), `X-Nexora-User` (optional principal), `Idempotency-Key`, `X-Request-Id`.

Errors use NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

| Area | Endpoints |
|------|-----------|
| Send | `POST /send`, `/send/bulk`, `/events` |
| Templates | `POST /templates`, `/templates/{id}/approve`, `/templates/preview` |
| Preferences | `GET|PUT /preferences/{principalId}` |
| Devices | `POST /devices` |
| Inbox | `GET /inbox/{principalId}`, `POST /inbox/{id}/read` |
| Deliveries | `GET /deliveries/{id}`, `POST .../retry`, `POST .../dlq` |
| Schedule | `POST /schedules`, `/schedules/{id}/cancel`, `/schedules/process-due` |
| AI stubs | `GET /ai/best-send-time/{principalId}`, `/ai/recommend-channel/{principalId}` |
| Admin | `GET /admin/stats`, `POST /outbox/publish` |

## Docs

- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [FEATURES.md](./FEATURES.md)
- OpenAPI: `api/openapi/notification-v1.yaml`
- Proto: `proto/notification/v1/notification.proto`
