# API Documentation Strategy

## Artifacts

| Surface | Spec | Publication |
|---------|------|-------------|
| BFF REST | OpenAPI 3.1 (YAML in repo) | Redoc / Stoplight per environment |
| Internal RPC | Protobuf + Buf | Generated stubs; breaking change CI |
| Async events | AsyncAPI for Kafka topics | Docs portal |
| Webhooks (future partners) | OpenAPI callbacks | Partner portal |

## Rules

1. Specs are committed **before or with** implementation — not after.
2. CI fails on OpenAPI/Protobuf breaking changes without explicit approval label.
3. Examples in specs use realistic NEXORA payloads (money as minor units).
4. Error codes catalog maintained in `docs/api/error-codes.md` (created in Phase 1).
5. Postman/Insomnia collections generated from OpenAPI — not hand-maintained as source of truth.

## Consumer compatibility

- Mobile min-version gates for breaking BFF changes
- Dual-run period documented in release notes
