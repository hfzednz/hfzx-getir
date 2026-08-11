# NEXORA CRM Service

Enterprise help desk: tickets, live chat, AI assistant (LLM port), knowledge base, cases, SLA/escalation, CSAT/NPS, Customer 360 aggregation.

**Module:** `github.com/nexora/crm-service`  
**HTTP:** `:8102` · **gRPC stub:** `:9102` · **Base path:** `/v1/crm/...`

## Hard boundaries

- Does **not** own notification delivery (`notification-service` via `NotificationClient`)
- Does **not** execute orders/payments/refunds (`RefundRequestClient` request-only)
- Does **not** own customer profile SoT (`ProfileReadClient` + `OrderReadClient` for Customer360)

AI bot orchestration currently lives behind the CRM `LLMClient` port (PROMPT-22). A dedicated `chat-assistant-service` may split later; see `../chat-assistant-service/README.md`.

## Quick start

```bash
# empty DATABASE_URL => in-memory mode
make run
# or
HTTP_ADDR=:8102 go run ./cmd/crm-service

make test
make build
```

Headers: `X-Tenant-Id` (required), `X-Nexora-User` (optional), `Idempotency-Key` (optional on create ticket).

## Environment

| Variable | Default | Notes |
|----------|---------|-------|
| `HTTP_ADDR` | `:8102` | REST listen |
| `GRPC_ADDR` | `:9102` | stub |
| `DATABASE_URL` | empty | empty = DevMode memory |
| `REDIS_URL` | empty | optional stub open |
| `KAFKA_BROKERS` | empty | stub publisher |
| `CORS_ALLOWED_ORIGINS` | `*` | CSV |
| `RATE_LIMIT_PER_MINUTE` | `240` | per IP |

## Layout

```text
cmd/crm-service/
internal/{config,domain,app,adapters,ratelimit}
migrations/ api/openapi/ proto/
```

## Key flows

1. Ticket: create → assign → resolve → close (close from `open` is rejected)
2. Chat: start → post message → transfer / end
3. AIAssist: intent + KB retrieve + LLM reply; escalate if low confidence or negative sentiment
4. SLA: policies set FirstResponseDue / ResolveDue; `EvaluateSLA` / `BreachEscalation`
5. Customer360: profile + orders stubs + local tickets/cases/CSAT

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).
