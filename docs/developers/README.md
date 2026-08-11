# NEXORA Developer Portal

API-first integration surface for partners and internal teams.

## Quick start

1. Create app: `POST /v1/open/apps`
2. Create API key: `POST /v1/open/apps/{id}/keys`
3. Browse catalog: `GET /v1/open/catalog?surface=public`
4. Register webhook: `POST /v1/open/webhooks`
5. Export gateway policies: `GET /v1/open/gateway/policies/export`

## Surfaces

| Surface | Audience | Examples |
|---------|----------|----------|
| public | 3P / mobile | orders, catalog, payments |
| private | admin/ops | ERP, warehouse, AI |
| partner | suppliers / marketplace | supplier, liveops, global |

## Auth

- OAuth2 / OIDC via `identity-service` (client id stored on app)
- API keys issued by open-platform (hashed at rest)
- Webhooks: `X-Nexora-Signature: sha256=...`

## SDKs

Published under `packages/sdk/{go,nodejs,python,flutter,...}` — thin clients over BFFs; no domain logic.

## Versioning

SemVer on catalog entries; deprecation recorded on `ApiVersion`. Migration guides live beside service OpenAPI.
